package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/api"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/db"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/mastery"
)

func main() {
	conn := db.Connect()

	for i := range 15 {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := conn.Ping(ctx)
		cancel()

		if err == nil {
			break
		}

		if i == 14 {
			log.Fatalf("database not reachable after retries: %v", err)
		}

		sleep := time.Duration(min(200*(1<<i), 60000)) * time.Millisecond
		time.Sleep(sleep)
	}
	
	defer conn.Close()

	service := mastery.NewMasteryService(conn)

	h := &api.Handler{Conn: conn, Service: service}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "https://tanaydonde.github.io"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/problems/{topic}", h.GetProblemsByTopic) // /api/problems/{topic}?handle=[handle]&inc=[inc]
		r.Get("/daily", h.GetDailyHandler)
		r.Get("/graph", h.GetGraphHandler)
		r.Get("/stats/{handle}", h.GetUserStats)
		r.Get("/recent/solved/{handle}", h.GetRecentSolvedHandler)
		r.Get("/recent/unsolved/{handle}", h.GetRecentUnsolvedHandler)
		r.Post("/sync/{handle}", h.SyncUserHandler)
		r.Post("/submit/{handle}", h.SubmitProblemHandler)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, r))
}
