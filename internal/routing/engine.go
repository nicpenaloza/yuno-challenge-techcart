package routing

import (
	"sort"
	"sync"
	"time"

	"github.com/yuno/techcart-failover/internal/domain"
	"github.com/yuno/techcart-failover/internal/health"
)

// Engine handles intelligent routing decisions
type Engine struct {
	mu         sync.RWMutex
	calculator *health.Calculator
	processors map[string]*domain.Processor
}

// NewEngine creates a new routing engine
func NewEngine(calc *health.Calculator) *Engine {
	return &Engine{
		calculator: calc,
		processors: make(map[string]*domain.Processor),
	}
}

// RegisterProcessor adds a processor configuration
func (e *Engine) RegisterProcessor(p *domain.Processor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.processors[p.ID] = p
}

// GetProcessors returns all registered processors
func (e *Engine) GetProcessors() []*domain.Processor {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*domain.Processor, 0, len(e.processors))
	for _, p := range e.processors {
		result = append(result, p)
	}
	return result
}

// Recommend returns ranked processors for a transaction scenario
func (e *Engine) Recommend(method domain.PaymentMethod, country domain.Country, amount float64) *domain.RoutingRecommendation {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Find candidates that support method + country
	candidates := e.findCandidates(method, country)

	// Rank by health
	rankings := e.rankProcessors(candidates)

	return &domain.RoutingRecommendation{
		Recommendations: rankings,
		PaymentMethod:   method,
		Country:         country,
		Timestamp:       time.Now(),
	}
}

// findCandidates returns processors supporting the method and country
func (e *Engine) findCandidates(method domain.PaymentMethod, country domain.Country) []*domain.Processor {
	var candidates []*domain.Processor

	for _, p := range e.processors {
		if e.supportsMethod(p, method) && e.supportsCountry(p, country) {
			candidates = append(candidates, p)
		}
	}

	return candidates
}

func (e *Engine) supportsMethod(p *domain.Processor, method domain.PaymentMethod) bool {
	for _, m := range p.PaymentMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (e *Engine) supportsCountry(p *domain.Processor, country domain.Country) bool {
	for _, c := range p.Countries {
		if c == country {
			return true
		}
	}
	return false
}

// rankProcessors ranks candidates by health status and auth rate
func (e *Engine) rankProcessors(processors []*domain.Processor) []domain.ProcessorRank {
	if len(processors) == 0 {
		return nil
	}

	type scored struct {
		processor *domain.Processor
		health    *domain.ProcessorHealth
		score     float64
	}

	scores := make([]scored, len(processors))
	for i, p := range processors {
		h := e.calculator.GetHealth(p.ID)
		scores[i] = scored{
			processor: p,
			health:    h,
			score:     e.calculateScore(h),
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Build rankings
	rankings := make([]domain.ProcessorRank, len(scores))
	for i, s := range scores {
		// Only recommend if HEALTHY or DEGRADED and first place
		recommended := i == 0 && s.health.Status != domain.StatusDown

		rankings[i] = domain.ProcessorRank{
			ProcessorID:       s.processor.ID,
			Rank:              i + 1,
			Status:            s.health.Status,
			AuthorizationRate: s.health.AuthorizationRate,
			Recommended:       recommended,
			Reason:            e.reasonForRank(s.health, i, recommended),
		}
	}

	return rankings
}

// calculateScore computes routing score for a processor
func (e *Engine) calculateScore(h *domain.ProcessorHealth) float64 {
	// Base score from auth rate (0-100)
	score := h.AuthorizationRate * 100

	// Penalize by status
	switch h.Status {
	case domain.StatusDown:
		score = 0 // Never route to DOWN
	case domain.StatusDegraded:
		score *= 0.5 // 50% penalty
	case domain.StatusHealthy:
		// No penalty
	}

	// Small bonus for more history (confidence)
	if h.TotalTransactions > 30 {
		score += 5
	}

	return score
}

// reasonForRank explains the ranking
func (e *Engine) reasonForRank(h *domain.ProcessorHealth, rank int, recommended bool) string {
	if h.Status == domain.StatusDown {
		return "Processor is DOWN - not recommended"
	}
	if h.Status == domain.StatusDegraded {
		if recommended {
			return "Best available but DEGRADED - use with caution"
		}
		return "DEGRADED - available as fallback"
	}
	if recommended {
		return "Best option - highest authorization rate"
	}
	return "Healthy fallback option"
}
