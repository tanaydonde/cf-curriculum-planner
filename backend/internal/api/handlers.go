package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/mastery"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/models"
)

type Handler struct {
    Conn *pgx.Conn
    Service *mastery.MasteryService
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

func (h *Handler) GetProblemsByTopic(w http.ResponseWriter, r *http.Request) {
	topic := chi.URLParam(r, "topic")
	handle := r.URL.Query().Get("handle")

	targetInc := 25
	if incStr := r.URL.Query().Get("inc"); incStr != "" {
		if val, err := strconv.Atoi(incStr); err == nil {
			targetInc = val
		}
	}

	limit := 5
	
	recommendations, err := h.Service.RecommendProblem(handle, topic, targetInc, limit)
	if err != nil {
		http.Error(w, "failed to get problems: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

func (h *Handler) SubmitProblemHandler(w http.ResponseWriter, r *http.Request) {
    handle := chi.URLParam(r, "handle")
    
    var input mastery.ProblemSolveInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.Service.UpdateSubmission(handle, input); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}