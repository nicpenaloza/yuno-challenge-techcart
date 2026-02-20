package health

import (
	"testing"
	"time"

	"github.com/yuno/techcart-failover/internal/domain"
)

// T2.1.1: Procesador nuevo → HEALTHY por defecto
func TestCalculator_NewProcessor_DefaultHealthy(t *testing.T) {
	calc := NewCalculator()

	health := calc.GetHealth("processor_a")

	if health.Status != domain.StatusHealthy {
		t.Errorf("expected HEALTHY, got %s", health.Status)
	}
	if health.AuthorizationRate != 1.0 {
		t.Errorf("expected 1.0 auth rate, got %f", health.AuthorizationRate)
	}
}

// T2.1.2: 80% approved → HEALTHY
func TestCalculator_80PercentApproved_Healthy(t *testing.T) {
	calc := NewCalculator()

	// Send 40 approved, 10 declined (within rolling window of 50)
	for i := 0; i < 40; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultApproved))
	}
	for i := 0; i < 10; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultDeclined))
	}

	health := calc.GetHealth("processor_a")

	if health.Status != domain.StatusHealthy {
		t.Errorf("expected HEALTHY, got %s", health.Status)
	}
	if health.AuthorizationRate < 0.79 || health.AuthorizationRate > 0.81 {
		t.Errorf("expected ~0.80 auth rate, got %f", health.AuthorizationRate)
	}
}

// T2.1.3: 50% approved → DEGRADED
func TestCalculator_50PercentApproved_Degraded(t *testing.T) {
	calc := NewCalculator()

	// Send 25 approved, 25 declined
	for i := 0; i < 25; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultApproved))
	}
	for i := 0; i < 25; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultDeclined))
	}

	health := calc.GetHealth("processor_a")

	if health.Status != domain.StatusDegraded {
		t.Errorf("expected DEGRADED, got %s", health.Status)
	}
}

// T2.1.4: 20% approved → DOWN
func TestCalculator_20PercentApproved_Down(t *testing.T) {
	calc := NewCalculator()

	// Send 10 approved, 40 declined
	for i := 0; i < 10; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultApproved))
	}
	for i := 0; i < 40; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultDeclined))
	}

	health := calc.GetHealth("processor_a")

	if health.Status != domain.StatusDown {
		t.Errorf("expected DOWN, got %s", health.Status)
	}
}

// T2.1.5: >50% errores → DOWN
func TestCalculator_HighErrorRate_Down(t *testing.T) {
	calc := NewCalculator()

	// Send 20 approved, 30 errors
	for i := 0; i < 20; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultApproved))
	}
	for i := 0; i < 30; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultError))
	}

	health := calc.GetHealth("processor_a")

	if health.Status != domain.StatusDown {
		t.Errorf("expected DOWN due to high error rate, got %s", health.Status)
	}
}

// T2.1.6: Rolling window respeta límite
func TestCalculator_RollingWindow_RespectsLimit(t *testing.T) {
	calc := NewCalculator()

	// Send 60 errors first (would make it DOWN)
	for i := 0; i < 60; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultError))
	}

	// Now send 50 approved (should push out old errors in rolling window)
	for i := 0; i < 50; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultApproved))
	}

	health := calc.GetHealth("processor_a")

	// With rolling window of 50, should now be HEALTHY
	if health.Status != domain.StatusHealthy {
		t.Errorf("expected HEALTHY after recovery, got %s", health.Status)
	}
}

// Test health transition is recorded
func TestCalculator_StatusTransition_RecordsAlert(t *testing.T) {
	calc := NewCalculator()

	// Start healthy
	for i := 0; i < 50; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultApproved))
	}

	// Now send errors to trigger DOWN
	for i := 0; i < 50; i++ {
		calc.RecordTransaction(createTx("processor_a", domain.ResultError))
	}

	transitions := calc.GetTransitions(time.Now().Add(-1 * time.Minute))

	if len(transitions) == 0 {
		t.Error("expected at least one transition")
	}
}

// Helper function
func createTx(processorID string, result domain.TransactionResult) domain.Transaction {
	return domain.Transaction{
		ID:            "test-tx",
		ProcessorID:   processorID,
		Timestamp:     time.Now(),
		Result:        result,
		PaymentMethod: domain.MethodPIX,
		Country:       domain.CountryBR,
		Amount:        100.0,
		Currency:      "BRL",
	}
}
