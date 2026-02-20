package health

import (
	"sync"
	"time"

	"github.com/yuno/techcart-failover/internal/domain"
)

// Configuration constants
const (
	WindowSize        = 50               // Rolling window: last N transactions
	TimeWindow        = 10 * time.Minute // Also consider time-based window
	HealthyThreshold  = 0.65             // >= 65% auth rate = HEALTHY
	DegradedThreshold = 0.30             // >= 30% auth rate = DEGRADED, below = DOWN
	ErrorRateDown     = 0.50             // > 50% error rate = DOWN
	ErrorRateDegraded = 0.30             // > 30% error rate = DEGRADED
	MinTransactions   = 10               // Min transactions before changing status
)

// Calculator tracks processor health based on transaction results
type Calculator struct {
	mu           sync.RWMutex
	transactions map[string][]domain.Transaction
	processors   map[string]*domain.ProcessorHealth
	transitions  []domain.HealthTransition
}

// NewCalculator creates a new health calculator
func NewCalculator() *Calculator {
	return &Calculator{
		transactions: make(map[string][]domain.Transaction),
		processors:   make(map[string]*domain.ProcessorHealth),
		transitions:  make([]domain.HealthTransition, 0),
	}
}

// RecordTransaction records a transaction and updates processor health
func (c *Calculator) RecordTransaction(tx domain.Transaction) *domain.ProcessorHealth {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add transaction
	c.transactions[tx.ProcessorID] = append(c.transactions[tx.ProcessorID], tx)

	// Prune old transactions
	c.pruneTransactions(tx.ProcessorID)

	// Recalculate health
	return c.calculateHealth(tx.ProcessorID)
}

// pruneTransactions keeps only relevant transactions
func (c *Calculator) pruneTransactions(processorID string) {
	txs := c.transactions[processorID]
	if len(txs) == 0 {
		return
	}

	cutoff := time.Now().Add(-TimeWindow)
	var recent []domain.Transaction

	for _, tx := range txs {
		if tx.Timestamp.After(cutoff) {
			recent = append(recent, tx)
		}
	}

	// Keep max 2x window size to have history
	if len(recent) > WindowSize*2 {
		recent = recent[len(recent)-WindowSize*2:]
	}

	c.transactions[processorID] = recent
}

// calculateHealth computes health status for a processor
func (c *Calculator) calculateHealth(processorID string) *domain.ProcessorHealth {
	txs := c.transactions[processorID]

	health := &domain.ProcessorHealth{
		ProcessorID: processorID,
		LastUpdated: time.Now(),
	}

	if len(txs) == 0 {
		health.Status = domain.StatusHealthy
		health.AuthorizationRate = 1.0
		c.processors[processorID] = health
		return health
	}

	// Use rolling window
	window := txs
	if len(window) > WindowSize {
		window = window[len(window)-WindowSize:]
	}

	var approved, declined, errors int
	for _, tx := range window {
		switch tx.Result {
		case domain.ResultApproved:
			approved++
		case domain.ResultDeclined:
			declined++
		case domain.ResultError, domain.ResultTimeout:
			errors++
		}
	}

	total := len(window)
	health.TotalTransactions = total
	health.SuccessCount = approved
	health.FailureCount = declined
	health.ErrorCount = errors

	// Calculate authorization rate: approved / (approved + declined)
	validAttempts := approved + declined
	if validAttempts > 0 {
		health.AuthorizationRate = float64(approved) / float64(validAttempts)
	} else if errors > 0 {
		health.AuthorizationRate = 0
	} else {
		health.AuthorizationRate = 1.0
	}

	// Calculate error rate
	errorRate := float64(errors) / float64(total)

	// Get previous status
	previousStatus := domain.StatusHealthy
	if prev, exists := c.processors[processorID]; exists {
		previousStatus = prev.Status
	}

	// Determine new status
	newStatus := c.determineStatus(health.AuthorizationRate, errorRate, total)
	health.Status = newStatus
	health.PreviousStatus = previousStatus

	// Record transition if changed
	if newStatus != previousStatus && c.processors[processorID] != nil {
		now := time.Now()
		health.StatusChangedAt = &now
		c.transitions = append(c.transitions, domain.HealthTransition{
			ProcessorID: processorID,
			FromStatus:  previousStatus,
			ToStatus:    newStatus,
			Timestamp:   now,
			Reason:      c.transitionReason(health.AuthorizationRate, errorRate),
		})
	}

	c.processors[processorID] = health
	return health
}

// determineStatus calculates health status based on rates
func (c *Calculator) determineStatus(authRate, errorRate float64, total int) domain.HealthStatus {
	// Need minimum transactions to change from default
	if total < MinTransactions {
		return domain.StatusHealthy
	}

	// High error rate = DOWN
	if errorRate > ErrorRateDown {
		return domain.StatusDown
	}

	// Elevated error rate = DEGRADED
	if errorRate > ErrorRateDegraded {
		return domain.StatusDegraded
	}

	// Low auth rate = DOWN
	if authRate < DegradedThreshold {
		return domain.StatusDown
	}

	// Medium auth rate = DEGRADED
	if authRate < HealthyThreshold {
		return domain.StatusDegraded
	}

	return domain.StatusHealthy
}

// transitionReason generates human-readable reason
func (c *Calculator) transitionReason(authRate, errorRate float64) string {
	if errorRate > ErrorRateDown {
		return "High error/timeout rate (>50%)"
	}
	if errorRate > ErrorRateDegraded {
		return "Elevated error/timeout rate (>30%)"
	}
	if authRate < DegradedThreshold {
		return "Very low authorization rate (<30%)"
	}
	if authRate < HealthyThreshold {
		return "Low authorization rate (<65%)"
	}
	return "Performance recovered"
}

// GetHealth returns current health for a processor
func (c *Calculator) GetHealth(processorID string) *domain.ProcessorHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if health, exists := c.processors[processorID]; exists {
		return health
	}

	return &domain.ProcessorHealth{
		ProcessorID:       processorID,
		Status:            domain.StatusHealthy,
		AuthorizationRate: 1.0,
		LastUpdated:       time.Now(),
	}
}

// GetAllHealth returns health for all tracked processors
func (c *Calculator) GetAllHealth() []*domain.ProcessorHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*domain.ProcessorHealth, 0, len(c.processors))
	for _, h := range c.processors {
		result = append(result, h)
	}
	return result
}

// GetTransitions returns health transitions (alerts) since given time
func (c *Calculator) GetTransitions(since time.Time) []domain.HealthTransition {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []domain.HealthTransition
	for _, t := range c.transitions {
		if t.Timestamp.After(since) {
			result = append(result, t)
		}
	}
	return result
}

// GetRecentTransactions returns recent transactions for a processor
func (c *Calculator) GetRecentTransactions(processorID string, limit int) []domain.Transaction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	txs := c.transactions[processorID]
	if len(txs) == 0 {
		return nil
	}

	if limit > 0 && len(txs) > limit {
		return txs[len(txs)-limit:]
	}
	return txs
}
