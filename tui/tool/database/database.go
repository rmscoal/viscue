package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"github.com/qustavo/sqlhooks/v2"

	sqlite3Migrator "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type sqlHook struct{}

const sqlHookContextKey = "sqlHook_query_executed_at"

func (*sqlHook) Before(
	ctx context.Context, query string, args ...interface{},
) (context.Context, error) {
	log.Debug("database sqlx hook", "query", query, "args", args)
	return context.WithValue(ctx, sqlHookContextKey, time.Now()), nil
}

func (*sqlHook) After(
	ctx context.Context, _ string, _ ...interface{},
) (context.Context, error) {
	begin := ctx.Value(sqlHookContextKey).(time.Time)
	log.Debug("database sqlx query took", "duration",
		time.Now().Sub(begin).Milliseconds())
	return ctx, nil
}

//go:embed migrations
var migrations embed.FS

func New() (*sqlx.DB, error) {
	sql.Register("sqlite3_with_sqlHook",
		sqlhooks.Wrap(&sqlite3.SQLiteDriver{}, &sqlHook{}))

	db, err := sqlx.Connect("sqlite3_with_sqlHook", "sqlite.db")
	if err != nil {
		err = fmt.Errorf("failed connecting to sqlite3: %s", err.Error())
	}

	driver, err := sqlite3Migrator.WithInstance(db.DB,
		&sqlite3Migrator.Config{})
	if err != nil {
		return nil, err
	}

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", source,
		"sqlite3_with_sqlHook", driver)
	if err != nil {
		return nil, err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed running migrations: %s", err.Error())
	}

	return db, nil
}
