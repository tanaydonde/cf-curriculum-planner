package api

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/jackc/pgx/v5"
    "github.com/tanaydonde/cf-curriculum-planner/backend/internal/db"
    "github.com/tanaydonde/cf-curriculum-planner/backend/internal/mastery"
    "github.com/tanaydonde/cf-curriculum-planner/backend/internal/models"
)

type Handler struct {
    Conn *pgx.Conn
    Service *mastery.MasteryService
}

func (h *Handler) GetProblemsByTopic(w http.ResponseWriter, r *http.Request) {
    topic := chi.URLParam(r, "topic")
    query := `SELECT problem_id, name, rating, tags FROM problems WHERE $1 = ANY(tags) LIMIT 50`
    
    rows, err := h.Conn.Query(r.Context(), query, topic)
    if err != nil {
        http.Error(w, "DB error", 500)
        return
    }
    defer rows.Close()

    var results []db.Problem
    for rows.Next() {
        var p db.Problem
        rows.Scan(&p.ID, &p.Name, &p.Rating, &p.Tags)
        results = append(results, p)
    }
    json.NewEncoder(w).Encode(results)
}

func (h *Handler) GetGraphHandler(w http.ResponseWriter, r *http.Request) {
    nodes, edges := models.GetGraph(h.Conn)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "nodes": nodes,
        "edges": edges,
    })
}

func (h *Handler) SyncUserHandler(w http.ResponseWriter, r *http.Request) {
    handle := chi.URLParam(r, "handle")
    if err := h.Service.Sync(handle); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.Write([]byte("Sync successful"))
}

func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
    handle := chi.URLParam(r, "handle")
    
    // This reuses your existing service logic
    stats, err := h.Service.RefreshAndGetAllStats(handle) 
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}