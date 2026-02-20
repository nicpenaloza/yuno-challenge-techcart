package main

import (
	"log"
	"net/http"
	"os"

	"github.com/yuno/techcart-failover/internal/api"
	"github.com/yuno/techcart-failover/internal/domain"
	"github.com/yuno/techcart-failover/internal/health"
	"github.com/yuno/techcart-failover/internal/routing"
)

func main() {
	// Initialize components
	calculator := health.NewCalculator()
	router := routing.NewEngine(calculator)

	// Register mock processors (TechCart scenario)
	registerProcessors(router)

	// Create API handler
	handler := api.NewHandler(calculator, router)

	// Setup routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Add CORS middleware for testing
	corsHandler := corsMiddleware(mux)

	// Start server (use PORT env var for Railway/cloud deployment)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("🚀 TechCart Failover API starting on %s", addr)
	log.Printf("📊 Registered %d processors", len(router.GetProcessors()))
	log.Println("")
	log.Println("Endpoints:")
	log.Println("  POST /api/v1/transactions     - Record transaction result")
	log.Println("  GET  /api/v1/health           - Get all processor health")
	log.Println("  GET  /api/v1/health/{id}      - Get processor health + history")
	log.Println("  POST /api/v1/routing/recommend - Get routing recommendation")
	log.Println("  GET  /api/v1/routing/recommend?payment_method=&country=")
	log.Println("  GET  /api/v1/processors       - List processors")
	log.Println("  GET  /api/v1/alerts           - Get health transitions")
	log.Println("")

	if err := http.ListenAndServe(addr, corsHandler); err != nil {
		log.Fatal(err)
	}
}

// registerProcessors sets up the mock processors for TechCart
func registerProcessors(router *routing.Engine) {
	processors := []*domain.Processor{
		{
			ID:             "processor_a",
			Name:           "GlobalPay_BR",
			Countries:      []domain.Country{domain.CountryBR},
			PaymentMethods: []domain.PaymentMethod{domain.MethodPIX, domain.MethodCard},
		},
		{
			ID:             "processor_b",
			Name:           "PayLatam",
			Countries:      []domain.Country{domain.CountryBR, domain.CountryMX, domain.CountryCO},
			PaymentMethods: []domain.PaymentMethod{domain.MethodCard},
		},
		{
			ID:             "processor_c",
			Name:           "PixMaster",
			Countries:      []domain.Country{domain.CountryBR},
			PaymentMethods: []domain.PaymentMethod{domain.MethodPIX},
		},
		{
			ID:             "processor_d",
			Name:           "MexPago",
			Countries:      []domain.Country{domain.CountryMX},
			PaymentMethods: []domain.PaymentMethod{domain.MethodCard, domain.MethodOXXO},
		},
		{
			ID:             "processor_e",
			Name:           "ColombiaPS",
			Countries:      []domain.Country{domain.CountryCO},
			PaymentMethods: []domain.PaymentMethod{domain.MethodPSE, domain.MethodCard},
		},
	}

	for _, p := range processors {
		router.RegisterProcessor(p)
	}
}

// corsMiddleware adds CORS headers for local testing
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
