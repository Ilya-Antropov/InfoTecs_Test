package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"Infotecs/internal/models"
)

var (
	ErrWalletNotFound      = errors.New("кошелек не найден")
	ErrInsufficientBalance = errors.New("недостаточно средств")
	ErrRecipientNotFound   = errors.New("кошелек получателя не найден")
)

const (
	createWalletsTableSQL = `
		CREATE TABLE IF NOT EXISTS wallets (
			address TEXT PRIMARY KEY,
			balance DECIMAL(10, 2) NOT NULL DEFAULT 0.0
		);`
	createTransactionsTableSQL = `
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			from_address TEXT NOT NULL,
			to_address TEXT NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
	countWalletsSQL           = "SELECT COUNT(*) FROM wallets"
	insertWalletSQL           = "INSERT INTO wallets (address, balance) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	selectBalanceForUpdateSQL = "SELECT balance FROM wallets WHERE address = $1 FOR UPDATE"
	updateSenderBalanceSQL    = "UPDATE wallets SET balance = balance - $1 WHERE address = $2"
	updateRecipientBalanceSQL = "UPDATE wallets SET balance = balance + $1 WHERE address = $2"
	insertTransactionSQL      = "INSERT INTO transactions (from_address, to_address, amount) VALUES ($1, $2, $3)"
	getTransactionsSQL        = `
		SELECT id, from_address, to_address, amount, created_at
		FROM transactions
		ORDER BY created_at DESC
		LIMIT $1`
	getWalletBalanceSQL = "SELECT balance FROM wallets WHERE address = $1"
)

type DB struct {
	*sql.DB
}

func InitDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть соединение с БД: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Close() {
	db.DB.Close()
}

func (db *DB) Initialize(ctx context.Context) error {
	if err := db.createTables(ctx); err != nil {
		return err
	}
	return db.createInitialWallets(ctx)
}

func (db *DB) createTables(ctx context.Context) error {
	_, err := db.ExecContext(ctx, createWalletsTableSQL)
	if err != nil {
		return fmt.Errorf("не удалось создать таблицу wallets: %w", err)
	}
	_, err = db.ExecContext(ctx, createTransactionsTableSQL)
	if err != nil {
		return fmt.Errorf("не удалось создать таблицу transactions: %w", err)
	}
	return nil
}

func generateWalletAddress() (string, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		return "", fmt.Errorf("не удалось сгенерировать случайные данные для адреса кошелька: %w", err)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (db *DB) createInitialWallets(ctx context.Context) error {
	const walletCount = 10
	const initialBalance = 100.0

	var count int
	row := db.QueryRowContext(ctx, countWalletsSQL)
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("не удалось подсчитать кошельки: %w", err)
	}

	if count >= walletCount {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию для создания кошельков: %w", err)
	}
	defer tx.Rollback()

	for i := 0; i < walletCount-count; i++ {
		addr, err := generateWalletAddress()
		if err != nil {
			return fmt.Errorf("не удалось сгенерировать адрес кошелька: %w", err)
		}

		if _, err := tx.ExecContext(
			ctx,
			insertWalletSQL,
			addr,
			initialBalance,
		); err != nil {
			return fmt.Errorf("не удалось вставить кошелек %s: %w", addr, err)
		}
	}
	return tx.Commit()
}

func (db *DB) SendMoney(ctx context.Context, from, to string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("сумма перевода должна быть положительной")
	}
	if from == to {
		return fmt.Errorf("нельзя переводить деньги самому себе")
	}

	txOptions := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}
	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback()

	var balance float64
	err = tx.QueryRowContext(ctx, selectBalanceForUpdateSQL, from).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWalletNotFound
		}
		return fmt.Errorf("ошибка при получении баланса отправителя: %w", err)
	}

	if balance < amount {
		return ErrInsufficientBalance
	}

	_, err = tx.ExecContext(ctx, updateSenderBalanceSQL, amount, from)
	if err != nil {
		return fmt.Errorf("не удалось обновить баланс отправителя: %w", err)
	}

	result, err := tx.ExecContext(ctx, updateRecipientBalanceSQL, amount, to)
	if err != nil {
		return fmt.Errorf("не удалось обновить баланс получателя: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRecipientNotFound
	}

	_, err = tx.ExecContext(
		ctx,
		insertTransactionSQL,
		from, to, amount,
	)
	if err != nil {
		return fmt.Errorf("не удалось записать транзакцию: %w", err)
	}

	return tx.Commit()
}

func (db *DB) GetTransactions(ctx context.Context, count int) ([]models.Transaction, error) {
	rows, err := db.QueryContext(ctx, getTransactionsSQL, count)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить транзакции: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err = rows.Scan(&t.ID, &t.FromAddress, &t.ToAddress, &t.Amount, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования транзакции: %w", err)
		}
		transactions = append(transactions, t)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по строкам: %w", err)
	}
	return transactions, nil
}

func (db *DB) GetWalletBalance(ctx context.Context, address string) (float64, error) {
	var balance float64
	err := db.QueryRowContext(ctx, getWalletBalanceSQL, address).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrWalletNotFound
		}
		return 0, fmt.Errorf("ошибка при получении баланса кошелька: %w", err)
	}
	return balance, nil
}
