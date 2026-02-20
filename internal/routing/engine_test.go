package routing

import (
	"testing"
	"time"

	"github.com/yuno/techcart-failover/internal/domain"
	"github.com/yuno/techcart-failover/internal/health"
)

func tx(processorID string, result domain.TransactionResult) domain.Transaction {
	return domain.Transaction{
		ProcessorID: processorID,
		Result:      result,
		Timestamp:   time.Now(),
	}
}

// T3.1.1: Recomendar procesador HEALTHY sobre DOWN
func TestEngine_RecommendHealthyOverDown(t *testing.T) {
	calc := health.NewCalculator()
	engine := NewEngine(calc)

	// Register processors
	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_a",
		Name:           "ProcessorA",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodPIX},
	})
	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_b",
		Name:           "ProcessorB",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodPIX},
	})

	// Make processor_a DOWN (lots of errors)
	for i := 0; i < 50; i++ {
		calc.RecordTransaction(tx("processor_a", domain.ResultError))
	}

	// Make processor_b HEALTHY
	for i := 0; i < 50; i++ {
		calc.RecordTransaction(tx("processor_b", domain.ResultApproved))
	}

	rec := engine.Recommend(domain.MethodPIX, domain.CountryBR, 100)

	if len(rec.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}

	// processor_b should be recommended (rank 1)
	if rec.Recommendations[0].ProcessorID != "processor_b" {
		t.Errorf("expected processor_b first, got %s", rec.Recommendations[0].ProcessorID)
	}
	if !rec.Recommendations[0].Recommended {
		t.Error("expected first processor to be recommended")
	}
}

// T3.1.2: Rankear por authorization_rate
func TestEngine_RankByAuthorizationRate(t *testing.T) {
	calc := health.NewCalculator()
	engine := NewEngine(calc)

	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_a",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodPIX},
	})
	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_b",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodPIX},
	})

	// processor_a: 70% auth rate
	for i := 0; i < 35; i++ {
		calc.RecordTransaction(tx("processor_a", domain.ResultApproved))
	}
	for i := 0; i < 15; i++ {
		calc.RecordTransaction(tx("processor_a", domain.ResultDeclined))
	}

	// processor_b: 90% auth rate
	for i := 0; i < 45; i++ {
		calc.RecordTransaction(tx("processor_b", domain.ResultApproved))
	}
	for i := 0; i < 5; i++ {
		calc.RecordTransaction(tx("processor_b", domain.ResultDeclined))
	}

	rec := engine.Recommend(domain.MethodPIX, domain.CountryBR, 100)

	if rec.Recommendations[0].ProcessorID != "processor_b" {
		t.Errorf("expected processor_b (90%%) first, got %s", rec.Recommendations[0].ProcessorID)
	}
}

// T3.1.3: Filtrar por payment_method
func TestEngine_FilterByPaymentMethod(t *testing.T) {
	calc := health.NewCalculator()
	engine := NewEngine(calc)

	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_pix",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodPIX},
	})
	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_card",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodCard},
	})

	rec := engine.Recommend(domain.MethodPIX, domain.CountryBR, 100)

	if len(rec.Recommendations) != 1 {
		t.Errorf("expected 1 recommendation, got %d", len(rec.Recommendations))
	}
	if rec.Recommendations[0].ProcessorID != "processor_pix" {
		t.Errorf("expected processor_pix, got %s", rec.Recommendations[0].ProcessorID)
	}
}

// T3.1.4: Filtrar por country
func TestEngine_FilterByCountry(t *testing.T) {
	calc := health.NewCalculator()
	engine := NewEngine(calc)

	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_br",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodCard},
	})
	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_mx",
		Countries:      []domain.Country{domain.CountryMX},
		PaymentMethods: []domain.PaymentMethod{domain.MethodCard},
	})

	rec := engine.Recommend(domain.MethodCard, domain.CountryMX, 100)

	if len(rec.Recommendations) != 1 {
		t.Errorf("expected 1 recommendation, got %d", len(rec.Recommendations))
	}
	if rec.Recommendations[0].ProcessorID != "processor_mx" {
		t.Errorf("expected processor_mx, got %s", rec.Recommendations[0].ProcessorID)
	}
}

// T3.1.5: Todos DOWN → no recommended
func TestEngine_AllDown_NoRecommended(t *testing.T) {
	calc := health.NewCalculator()
	engine := NewEngine(calc)

	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_a",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodPIX},
	})
	engine.RegisterProcessor(&domain.Processor{
		ID:             "processor_b",
		Countries:      []domain.Country{domain.CountryBR},
		PaymentMethods: []domain.PaymentMethod{domain.MethodPIX},
	})

	// Make both DOWN
	for i := 0; i < 50; i++ {
		calc.RecordTransaction(tx("processor_a", domain.ResultError))
		calc.RecordTransaction(tx("processor_b", domain.ResultError))
	}

	rec := engine.Recommend(domain.MethodPIX, domain.CountryBR, 100)

	// Should have results but none recommended
	for _, r := range rec.Recommendations {
		if r.Recommended {
			t.Errorf("expected no processor to be recommended when all DOWN")
		}
	}
}
