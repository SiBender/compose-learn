package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbName = "compose-test-db"
	query  = "SELECT data FROM data_table ORDER BY id DESC LIMIT 1 OFFSET 0"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	port := getEnv("PORT", "8080")
	dbConn := getEnv("DB_CONNECTION", "localhost:5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")

	dsn := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(dbUser, dbPassword),
		Host:     dbConn,
		Path:     "/" + dbName,
		RawQuery: "sslmode=disable",
	}).String()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Printf("warning: database is not reachable yet: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var data string
		err := pool.QueryRow(r.Context(), query).Scan(&data)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "no data", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("query failed: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, data)
	})

	log.Printf("listening on :%s (db %s/%s)", port, dbConn, dbName)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
