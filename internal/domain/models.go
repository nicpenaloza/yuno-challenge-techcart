package domain

import "time"

// TransactionResult represents the outcome of a payment transaction
type TransactionResult string

const (
	ResultApproved TransactionResult = "approved"
	ResultDeclined TransactionResult = "declined"
	ResultError    TransactionResult = "error"
	ResultTimeout  TransactionResult = "timeout"
)

// HealthStatus represents the health state of a processor
type HealthStatus string

const (
	StatusHealthy  HealthStatus = "HEALTHY"
	StatusDegraded HealthStatus = "DEGRADED"
	StatusDown     HealthStatus = "DOWN"
)

// PaymentMethod represents supported payment methods
type PaymentMethod string

const (
	MethodPIX  PaymentMethod = "PIX"
	MethodCard PaymentMethod = "CARD"
	MethodOXXO PaymentMethod = "OXXO"
	MethodPSE  PaymentMethod = "PSE"
)

// Country represents supported countries
type Country string

const (
	CountryBR Country = "BR"
	CountryMX Country = "MX"
	CountryCO Country = "CO"
)

// Transaction represents a transaction result received from merchants
type Transaction struct {
	ID            string            `json:"id"`
	ProcessorID   string            `json:"processor_id"`
	Timestamp     time.Time         `json:"timestamp"`
	Result        TransactionResult `json:"result"`
	PaymentMethod PaymentMethod     `json:"payment_method"`
	Country       Country           `json:"country"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
}

// Processor represents a payment processor configuration
type Processor struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Countries      []Country       `json:"countries"`
	PaymentMethods []PaymentMethod `json:"payment_methods"`
}

// ProcessorHealth represents the current health state of a processor
type ProcessorHealth struct {
	ProcessorID       string       `json:"processor_id"`
	Status            HealthStatus `json:"status"`
	AuthorizationRate float64      `json:"authorization_rate"`
	TotalTransactions int          `json:"total_transactions"`
	SuccessCount      int          `json:"success_count"`
	FailureCount      int          `json:"failure_count"`
	ErrorCount        int          `json:"error_count"`
	LastUpdated       time.Time    `json:"last_updated"`
	StatusChangedAt   *time.Time   `json:"status_changed_at,omitempty"`
	PreviousStatus    HealthStatus `json:"previous_status,omitempty"`
}

// RoutingRecommendation represents the routing decision
type RoutingRecommendation struct {
	Recommendations []ProcessorRank `json:"recommendations"`
	PaymentMethod   PaymentMethod   `json:"payment_method"`
	Country         Country         `json:"country"`
	Timestamp       time.Time       `json:"timestamp"`
}

// ProcessorRank represents a processor's ranking for routing
type ProcessorRank struct {
	ProcessorID       string       `json:"processor_id"`
	Rank              int          `json:"rank"`
	Status            HealthStatus `json:"status"`
	AuthorizationRate float64      `json:"authorization_rate"`
	Recommended       bool         `json:"recommended"`
	Reason            string       `json:"reason"`
}

// HealthTransition records when a processor changes health status
type HealthTransition struct {
	ProcessorID string       `json:"processor_id"`
	FromStatus  HealthStatus `json:"from_status"`
	ToStatus    HealthStatus `json:"to_status"`
	Timestamp   time.Time    `json:"timestamp"`
	Reason      string       `json:"reason"`
}
