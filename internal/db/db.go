package DB

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/lib/pq"
)

var (
	Conn    *pgx.Conn
	Queries *sqlc.Queries
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
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbConfig, nil := pgxpool.ParseConfig(fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", opts.Host, opts.Port, opts.User, opts.Name, opts.Password, opts.Sslmode))
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, dbConfig)
	// conn, err := pgx.Connect(ctx, fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", opts.Host, opts.Port, opts.User, opts.Name, opts.Password, opts.Sslmode))
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}

	customTypes, err := getCustomDataTypes(context.Background(), conn)
	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, t := range customTypes {
			conn.TypeMap().RegisterType(t)
		}
		return nil
	}
	// Immediately close the old pool and open a new one with the new dbConfig.
	conn.Close()
	conn, err = pgxpool.NewWithConfig(context.Background(), dbConfig)

	Queries = sqlc.New(conn)
}

func ShutdownConnection() error {
	if Conn.IsClosed() {
		return nil
	}
	fmt.Print("Shutting down...")
	ctx := context.Background()
	err := Conn.Close(ctx)
	if err != nil {
		return err
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
	defer conn.Release()
	if err != nil {
		return nil, err
	}

	// TODO: Add missing custom types
	dataTypeNames := []string{
		"wettkampf",
		// An underscore prefix is an array type in pgtypes.
		"_wettkampf",
	}

	var typesToRegister []*pgtype.Type
	for _, typeName := range dataTypeNames {
		dataType, err := conn.Conn().LoadType(ctx, typeName)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to load type %s: %v", typeName, err))
		}
		// You need to register only for this connection too, otherwise the array type will look for the register element type.
		conn.Conn().TypeMap().RegisterType(dataType)
		typesToRegister = append(typesToRegister, dataType)
	}
	return typesToRegister, nil
}
