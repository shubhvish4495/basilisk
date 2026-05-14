package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"basilisk/pkg/helper"
)

var instance DB

// Config holds the database connection configuration parameters.
type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl"`
}

// DB defines the interface for database operations.
type DB interface {
	// default methods for DB interface
	Ping() error
	PingContext(ctx context.Context) error
	Close() error

	// jwt auth required db methods
	GetUserByEmail(ctx context.Context, logger *slog.Logger, emailID string) (*User, helper.HttpError)
	GetUserByID(ctx context.Context, logger *slog.Logger, userID string) (*User, helper.HttpError)
	CreateUser(ctx context.Context, logger *slog.Logger, user User) (uuid.UUID, helper.HttpError)
}

// DBStruct wraps sql.DB and implements the DB interface.
type DBStruct struct {
	*sql.DB
}

// getConnectionString constructs a PostgreSQL connection string from the provided configuration.
// It formats the configuration parameters into a DSN (Data Source Name) that can be used
// to connect to a PostgreSQL database. If SSLMode is not specified, it defaults to "disable".
//
// Parameters:
//   - c: The database configuration containing connection parameters
//
// Returns:
//   - string: A formatted PostgreSQL connection string
func getConnectionString(c Config) string {
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		c.Host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode)
}

// GetInstance returns the singleton database instance.
// This function provides access to the initialized database instance that was
// created during the Init call. It should only be called after successful initialization.
//
// Returns:
//   - DB: The initialized database instance implementing the DB interface
func GetInstance() DB {
	return instance
}

// SetInstance overrides the global database instance. It is intended for use in tests only.
func SetInstance(db DB) {
	instance = db
}

// Init initializes the database connection and sets up the global database instance.
// It establishes a connection to PostgreSQL using the provided configuration
// and verifies connectivity with a ping.
//
// Parameters:
//   - ctx: Context for the database ping operation
//   - c: Database configuration containing connection parameters
//
// Returns:
//   - error: An error if database connection or ping fails, nil otherwise
func Init(ctx context.Context, c Config) error {
	dsn := getConnectionString(c)

	// open sql connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	// ping database to check connection
	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	// set instance of type DBStruct
	instance = &DBStruct{
		DB: db,
	}
	return nil
}

// withTransaction executes the given function within a database transaction.
// It automatically handles beginning, committing, and rolling back the transaction.
// If the function returns an error, the transaction is rolled back; otherwise, it is committed.
//
// Parameters:
//   - ctx: Context for the transaction
//   - logger: Logger for recording transaction errors
//   - fn: The function to execute within the transaction scope
//
// Returns:
//   - error: An error if the transaction fails to begin, commit, or rollback, nil otherwise
func (db *DBStruct) withTransaction(ctx context.Context, logger *slog.Logger, fn func(logger *slog.Logger, tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		logger.Error("error while starting transaction", "error", err)
		return err
	}

	if err = fn(logger, tx); err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			logger.Error("error while rolling back the transaction", "error", rbErr)
		}
		return errors.Join(err, rbErr)
	}

	if err = tx.Commit(); err != nil {
		logger.Error("error while committing transaction", "error", err)
		return err
	}

	return nil
}

func (db *DBStruct) insertReturningID(ctx context.Context, logger *slog.Logger, query string, args ...any) (uuid.UUID, helper.HttpError) {
	logger.Debug("running query to insert into database", "query", query)

	var id uuid.UUID
	err := db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		logger.Error("error while inserting into database", "error", err)
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Unique violation
				return uuid.Nil, helper.NewHttpError(helper.ConflictError.Code, "data already exist in db")
			}
		}
		return uuid.Nil, helper.InternalServerError
	}

	return id, nil
}

func (db *DBStruct) getSingleRow(ctx context.Context, logger *slog.Logger, query string, args ...any) (*sql.Row, helper.HttpError) {
	logger.Debug("running query to get data from database", "query", query)

	row := db.QueryRowContext(ctx, query, args...)
	if row.Err() != nil {
		logger.Error("error while querying in row from database", "error", row.Err().Error())
		return nil, helper.InternalServerError
	}

	return row, nil
}

func (db *DBStruct) DummyTxMethod(ctx context.Context, logger *slog.Logger) {
	_ = db.withTransaction(ctx, logger, func(logger *slog.Logger, tx *sql.Tx) error {
		return nil
	})
}
