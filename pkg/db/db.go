package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

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
	GetUser(ctx context.Context, userID string) (*User, helper.HttpError)
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
