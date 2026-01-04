package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/api"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/db"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/mastery"
)

func test(service *mastery.MasteryService) {
	//service.Sync("tourist")
	recProblems, _ := service.RecommendProblem("tanay5", "greedy", 300, 5)
	for _, problem := range recProblems {
		fmt.Printf("[%d] %s (%s) Tags: %v\n", problem.Rating, problem.Name, problem.ID, problem.Tags)
	}
}

func main() {
	testing := false
	conn := db.Connect()
	defer conn.Close(context.Background())

	service := mastery.NewMasteryService(conn)

	if testing{
		test(service)
		fmt.Println("test succeeded")
		return
	}

	h := &api.Handler{Conn: conn, Service: service,}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"}, // Your frontend URLs
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/problems/{topic}", h.GetProblemsByTopic) // /api/problems/{topic}?handle=[handle]&inc=[inc]
		r.Get("/graph", h.GetGraphHandler)
		r.Get("/stats/{handle}", h.GetUserStats)
		r.Post("/sync/{handle}", h.SyncUserHandler)
		r.Post("/submit/{handle}", h.SubmitProblemHandler)
	})

	port := ":8080"
	fmt.Printf("Server starting on %s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
}
