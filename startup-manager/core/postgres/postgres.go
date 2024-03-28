package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"startup-manager/core/config"
)

type Postgres struct {
	DB *sqlx.DB
}

const (
	connectAttempts = 15
)

// NewPostgres creates a new Postgres instance
func NewPostgres(cfg *core.DbConfig) (Postgres, error) {
	var (
		db  *sql.DB
		m   *migrate.Migrate
		err error
		d   database.Driver
	)
	// log.Println("new postgres called")
	// log.Println("cfg********", cfg)
	// log.Println(cfg.ConnectionString, "connection string")
	db, err = sql.Open("postgres", cfg.ConnectionString)
	if err != nil {

		log.Println(err, "error in db")
	}
	// for i := 0; i < connectAttempts; i++ {
	// 	time.Sleep(time.Second)
	// 	log.Println(cfg.ConnectionString)

	// }
	d, err = postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to create postgres driver: %w", err)
	}
	// log.Println(cfg.MigrationsPath, "Migration path")
	fs, err := os.ReadDir(cfg.MigrationsPath)
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to stat migrations path: %w", err)
	}

	if len(fs) == 0 {
		return Postgres{}, fmt.Errorf("no migrations found in %s", cfg.MigrationsPath)
	}

	for _, f := range fs {
		if !f.IsDir() {
			continue
		}

		fmt.Println(f.Name())
	}

	migrationsPath := fmt.Sprintf("file://%s", cfg.MigrationsPath)
	m, err = migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres", d)
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to create migrations instance: %w", err)
	}
	// 	err = m.Force(2) // Set 2 to the desired version
	// if err != nil {
	//     return Postgres{}, fmt.Errorf("failed to force set database version: %w", err)
	// }
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return Postgres{}, fmt.Errorf("failed to apply migrations: %w", err)
	}

	sdb := sqlx.NewDb(db, "postgres")
	sdb.SetMaxOpenConns(cfg.MaxOpenConns)

	err = sdb.Ping()
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to ping database: %w", err)
	}

	return Postgres{
		DB: sdb,
	}, nil
}

func (p Postgres) Name() string {
	return "postgres db"
}

func (p Postgres) CheckHealth(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

func (p Postgres) CheckReadiness(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

// Close closes the database connection
func (p *Postgres) Close() error {
	return p.DB.Close()
}
