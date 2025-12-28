package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/jackc/pgx/v5"
)

var conn *pgx.Conn

func main() {
	var err error

	connStr := "postgres://donde783985@localhost:5432/cf_planner"

	conn, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot connect to database: %v\n", err)
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	script, err := os.ReadFile("init.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read SQL file: %v\n", err)
		os.Exit(1)
	}

	_, err = conn.Exec(context.Background(), string(script))
	if err != nil {
		fmt.Fprintf(os.Stderr, "SQL execution failed: %v\n", err)
		os.Exit(1)
	}

	// problems, err := getProblems()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// createTables(problems)

	// fmt.Println("database tables initialized")

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"}, 
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: true,
		MaxAge: 300, 
	}))

	r.Use(middleware.Logger)

	r.Get("/problems/{topic}", getProblemsByTopic)
	r.Get("/graph", getGraphHandler)

	fmt.Println("Server starting on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}