package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/yuno/techcart-failover/internal/domain"
	"github.com/yuno/techcart-failover/internal/health"
	"github.com/yuno/techcart-failover/internal/routing"
)

// Handler holds API dependencies
type Handler struct {
	calculator *health.Calculator
	router     *routing.Engine
}

// NewHandler creates a new API handler
func NewHandler(calc *health.Calculator, router *routing.Engine) *Handler {
	return &Handler{
		calculator: calc,
		router:     router,
	}
}

// Request/Response types

type TransactionRequest struct {
	ProcessorID   string  `json:"processor_id"`
	Result        string  `json:"result"`
	PaymentMethod string  `json:"payment_method"`
	Country       string  `json:"country"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Timestamp     string  `json:"timestamp,omitempty"`
}

type RoutingRequest struct {
	PaymentMethod string  `json:"payment_method"`
	Country       string  `json:"country"`
	Amount        float64 `json:"amount"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Transaction recording
	mux.HandleFunc("POST /api/v1/transactions", h.RecordTransaction)

	// Health monitoring
	mux.HandleFunc("GET /api/v1/health", h.GetAllHealth)
	mux.HandleFunc("GET /api/v1/health/{processorId}", h.GetProcessorHealth)

	// Routing
	mux.HandleFunc("POST /api/v1/routing/recommend", h.GetRoutingRecommendation)
	mux.HandleFunc("GET /api/v1/routing/recommend", h.GetRoutingRecommendationQuery)

	// Processors
	mux.HandleFunc("GET /api/v1/processors", h.GetProcessors)

	// Alerts
	mux.HandleFunc("GET /api/v1/alerts", h.GetAlerts)
}

// POST /api/v1/transactions - Record a transaction result
func (h *Handler) RecordTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProcessorID == "" {
		h.writeError(w, "processor_id is required", http.StatusBadRequest)
		return
	}

	// Parse timestamp or use current time
	timestamp := time.Now()
	if req.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			timestamp = t
		}
	}

	tx := domain.Transaction{
		ID:            generateID(),
		ProcessorID:   req.ProcessorID,
		Timestamp:     timestamp,
		Result:        domain.TransactionResult(req.Result),
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		Country:       domain.Country(req.Country),
		Amount:        req.Amount,
		Currency:      req.Currency,
	}

	health := h.calculator.RecordTransaction(tx)
	h.writeJSON(w, health, http.StatusOK)
}

// GET /api/v1/health - Get health status of all processors
func (h *Handler) GetAllHealth(w http.ResponseWriter, r *http.Request) {
	health := h.calculator.GetAllHealth()
	h.writeJSON(w, map[string]interface{}{
		"processors": health,
		"count":      len(health),
		"timestamp":  time.Now(),
	}, http.StatusOK)
}

// GET /api/v1/health/{processorId} - Get health + history for a processor
func (h *Handler) GetProcessorHealth(w http.ResponseWriter, r *http.Request) {
	processorID := r.PathValue("processorId")
	if processorID == "" {
		h.writeError(w, "processor_id is required", http.StatusBadRequest)
		return
	}

	health := h.calculator.GetHealth(processorID)
	recentTxs := h.calculator.GetRecentTransactions(processorID, 20)

	h.writeJSON(w, map[string]interface{}{
		"health":              health,
		"recent_transactions": recentTxs,
		"transaction_count":   len(recentTxs),
	}, http.StatusOK)
}

// POST /api/v1/routing/recommend - Get routing recommendation
func (h *Handler) GetRoutingRecommendation(w http.ResponseWriter, r *http.Request) {
	var req RoutingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PaymentMethod == "" || req.Country == "" {
		h.writeError(w, "payment_method and country are required", http.StatusBadRequest)
		return
	}

	recommendation := h.router.Recommend(
		domain.PaymentMethod(req.PaymentMethod),
		domain.Country(req.Country),
		req.Amount,
	)
	h.writeJSON(w, recommendation, http.StatusOK)
}

// GET /api/v1/routing/recommend?payment_method=&country=&amount=
func (h *Handler) GetRoutingRecommendationQuery(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("payment_method")
	country := r.URL.Query().Get("country")

	if method == "" || country == "" {
		h.writeError(w, "payment_method and country query params are required", http.StatusBadRequest)
		return
	}

	recommendation := h.router.Recommend(
		domain.PaymentMethod(method),
		domain.Country(country),
		0,
	)
	h.writeJSON(w, recommendation, http.StatusOK)
}

// GET /api/v1/processors - List all registered processors
func (h *Handler) GetProcessors(w http.ResponseWriter, r *http.Request) {
	processors := h.router.GetProcessors()
	h.writeJSON(w, map[string]interface{}{
		"processors": processors,
		"count":      len(processors),
	}, http.StatusOK)
}

// GET /api/v1/alerts - Get health status transitions
func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	// Default to last hour
	since := time.Now().Add(-1 * time.Hour)

	if sinceParam := r.URL.Query().Get("since"); sinceParam != "" {
		if t, err := time.Parse(time.RFC3339, sinceParam); err == nil {
			since = t
		}
	}

	transitions := h.calculator.GetTransitions(since)
	h.writeJSON(w, map[string]interface{}{
		"alerts":    transitions,
		"count":     len(transitions),
		"since":     since,
		"timestamp": time.Now(),
	}, http.StatusOK)
}

// Helper methods

func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, message string, status int) {
	h.writeJSON(w, ErrorResponse{Error: message}, status)
}

var idCounter int64

func generateID() string {
	id := atomic.AddInt64(&idCounter, 1)
	return fmt.Sprintf("tx-%d-%d", time.Now().UnixNano(), id)
}
