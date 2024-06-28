package DB

import (
	"context"
	"fmt"
	"time"

	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	// "github.com/jackc/pgx/v5/pgxpool"

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

	// conn, err := pgxpool.New(ctx, fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", opts.Host, opts.Port, opts.User, opts.Name, opts.Password, opts.Sslmode))
	conn, err := pgx.Connect(ctx, fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", opts.Host, opts.Port, opts.User, opts.Name, opts.Password, opts.Sslmode))
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
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
