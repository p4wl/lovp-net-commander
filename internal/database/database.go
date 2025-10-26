package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	app "net-commander-server/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Username string
	Password string
	Host     string
	Port     uint16
	DBName   string
	SSLMode  string
}

func NewDatabase(appCfg *app.AppCfg) (*Database, error) {
	db := &Database{
		Username: appCfg.Database.User,
		Password: appCfg.Database.Password,
		Host:     appCfg.Database.Host,
		Port:     appCfg.Database.Port,
		DBName:   appCfg.Database.Name,
		SSLMode:  appCfg.Database.SSLMode,
	}

	err := db.validate()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *Database) validate() error {
	if db.Username == "" {
		return fmt.Errorf("invalid database.username")
	}
	if db.Password == "" {
		return fmt.Errorf("invalid database.Password")
	}
	if db.Host == "" {
		return fmt.Errorf("invalid database.Host")
	}
	if db.Port == 0 {
		return fmt.Errorf("invalid database.Port")
	}
	if db.DBName == "" {
		return fmt.Errorf("invalid database.DBName")
	}
	if db.SSLMode == "" {
		return fmt.Errorf("invalid database.SSLMode")
	}

	return nil
}

func (c *Database) URL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.SSLMode,
	)
}

func Connect(ctx context.Context, logger *slog.Logger, migrations fs.FS, appCfg *app.AppCfg) (*pgxpool.Pool, error) {
	db, err := NewDatabase(appCfg)
	if err != nil {
		return nil, err
	}

	pgCfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		db.Username, db.Password, db.Host, db.Port, db.DBName, db.SSLMode,
	))
	if err != nil {
		return nil, err
	}

	connection, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, err
	}

	url := db.URL()

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, err
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", source, url)
	if err != nil {
		return nil, err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return connection, nil
}
