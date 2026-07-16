package DB

import (
	"context"
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

// Any custom DB types made with CREATE TYPE need to be registered with pgx.
// https://github.com/kyleconroy/sqlc/issues/2116
// https://stackoverflow.com/questions/75658429/need-to-update-psql-row-of-a-composite-type-in-golang-with-jack-pgx
// https://pkg.go.dev/github.com/jackc/pgx/v5/pgtype
func getCustomDataTypes(ctx context.Context, pool *pgxpool.Pool) ([]*pgtype.Type, error) {
	// Get a single connection just to load type information.
	conn, err := pool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, err
	}

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
