package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"UASBE/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDB(cfg config.Config) *pgxpool.Pool {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB,
	)


	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("❌ Failed parsing PostgreSQL config: %v", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("❌ Failed connect PostgreSQL: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("❌ PostgreSQL unreachable: %v", err)
	}

	log.Println("✅ PostgreSQL connected successfully")
	return db
}
