package migrations

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"gopricemon/internal/repository/postgres"
)

func Up(dsn string) error {
	return run(dsn, func(db *sql.DB, dir string) error {
		return goose.Up(db, dir)
	})
}

func Down(dsn string) error {
	return run(dsn, func(db *sql.DB, dir string) error {
		return goose.Down(db, dir)
	})
}

func run(dsn string, migrate func(*sql.DB, string) error) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	goose.SetBaseFS(postgres.Migrations)
	defer goose.SetBaseFS(nil)

	return migrate(db, postgres.MigrationsDir)
}
