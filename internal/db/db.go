package DB

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	Queries *sqlc.Queries
	pool    *pgxpool.Pool
)

type DBServerOptions struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Sslmode  string
}

func InitConnection(opts DBServerOptions) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		opts.Host, opts.Port, opts.User, opts.Name, opts.Password, opts.Sslmode)

	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to parse db config: %v", err))
		os.Exit(1)
	}

	pool, err = pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		slog.Error(fmt.Sprintf("failed opening connection to postgres: %v", err))
		os.Exit(1)
	}

	customTypes, err := getCustomDataTypes(ctx, pool)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to load custom pg types: %v", err))
		os.Exit(1)
	}

	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, t := range customTypes {
			conn.TypeMap().RegisterType(t)
		}
		return nil
	}

	pool.Close()
	pool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		slog.Error(fmt.Sprintf("failed opening connection to postgres: %v", err))
		os.Exit(1)
	}

	Queries = sqlc.New(pool)
}

func ShutdownConnection() error {
	if pool == nil {
		return nil
	}
	slog.Info("Shutting down database connection...")
	pool.Close()
	return nil
}

type queriesCtxKey struct{}

// WithQueries returns a context carrying tx-bound queries. crud functions
// pick these up via QueriesFromCtx so they run within an open transaction.
func WithQueries(ctx context.Context, q *sqlc.Queries) context.Context {
	return context.WithValue(ctx, queriesCtxKey{}, q)
}

// QueriesFromCtx returns the queries bound to ctx, defaulting to the global
// pool-bound Queries when no transaction is active.
func QueriesFromCtx(ctx context.Context) *sqlc.Queries {
	if q, ok := ctx.Value(queriesCtxKey{}).(*sqlc.Queries); ok && q != nil {
		return q
	}
	return Queries
}

// WithTx runs fn within a single transaction. It is reentrant: if ctx already
// carries tx-bound queries, fn runs on that existing transaction instead of
// starting a nested one.
func WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if _, ok := ctx.Value(queriesCtxKey{}).(*sqlc.Queries); ok {
		return fn(ctx)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		rbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if rbErr := tx.Rollback(rbCtx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.Warn("transaction rollback failed", "err", rbErr)
		}
	}()

	q := Queries.WithTx(tx)
	txCtx := WithQueries(ctx, q)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// Any custom DB types made with CREATE TYPE need to be registered with pgx.
// https://github.com/kyleconroy/sqlc/issues/2116
// https://stackoverflow.com/questions/75658429/need-to-update-psql-row-of-a-composite-type-in-golang-with-jack-pgx
// https://pkg.go.dev/github.com/jackc/pgx/v5/pgtype
func getCustomDataTypes(ctx context.Context, pool *pgxpool.Pool) ([]*pgtype.Type, error) {
	// Get a single connection just to load type information.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// TODO: Add missing custom types
	dataTypeNames := []string{
		"wettkampf",
		"_wettkampf",
		"user_capability",
		"_user_capability",
		"geschlecht",
		"_geschlecht",
		"tag",
		"_tag",
		"rolle",
		"_rolle",
	}

	var typesToRegister []*pgtype.Type
	for _, typeName := range dataTypeNames {
		dataType, err := conn.Conn().LoadType(ctx, typeName)
		if err != nil {
			return nil, fmt.Errorf("failed to load type %s: %v", typeName, err)
		}
		// You need to register only for this connection too, otherwise the array type will look for the register element type.
		conn.Conn().TypeMap().RegisterType(dataType)
		typesToRegister = append(typesToRegister, dataType)
	}
	return typesToRegister, nil
}
