# TechCart Failover Intelligence API

Smart payment routing engine with real-time processor health monitoring. Built for TechCart to prevent revenue loss during payment processor outages.

## Problem Solved

TechCart lost $420,000 during a 4-hour processor outage because their static routing couldn't detect and route around the failure. This API provides:

- **Real-time health monitoring** - Detects processor failures within seconds
- **Intelligent routing** - Automatically routes to healthy processors
- **Operational visibility** - Alerts and dashboards for the ops team

## Quick Start

```bash
# Start the server
go run cmd/server/main.go

# Run the demo (in another terminal)
./scripts/demo.sh
```

The demo simulates a realistic outage scenario with 1100+ transactions.

## API Endpoints

### Record Transaction Result

```bash
POST /api/v1/transactions

curl -X POST localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "processor_id": "processor_a",
    "result": "approved",
    "payment_method": "PIX",
    "country": "BR",
    "amount": 150.00,
    "currency": "BRL"
  }'
```

**Results:** `approved`, `declined`, `error`, `timeout`

### Get All Processor Health

```bash
GET /api/v1/health

curl localhost:8080/api/v1/health | jq
```

Response:
```json
{
  "processors": [
    {
      "processor_id": "processor_a",
      "status": "HEALTHY",
      "authorization_rate": 0.78,
      "total_transactions": 50,
      "success_count": 39,
      "failure_count": 8,
      "error_count": 3
    }
  ],
  "count": 5,
  "timestamp": "2024-02-20T12:30:00Z"
}
```

### Get Processor Health + History

```bash
GET /api/v1/health/{processorId}

curl localhost:8080/api/v1/health/processor_a | jq
```

### Get Routing Recommendation

```bash
# Query params
GET /api/v1/routing/recommend?payment_method=PIX&country=BR

curl 'localhost:8080/api/v1/routing/recommend?payment_method=PIX&country=BR' | jq

# Or POST with JSON
POST /api/v1/routing/recommend
{
  "payment_method": "PIX",
  "country": "BR",
  "amount": 100.00
}
```

Response:
```json
{
  "recommendations": [
    {
      "processor_id": "processor_c",
      "rank": 1,
      "status": "HEALTHY",
      "authorization_rate": 0.82,
      "recommended": true,
      "reason": "Best option - highest authorization rate"
    },
    {
      "processor_id": "processor_a",
      "rank": 2,
      "status": "DOWN",
      "authorization_rate": 0.10,
      "recommended": false,
      "reason": "Processor is DOWN - not recommended"
    }
  ],
  "payment_method": "PIX",
  "country": "BR"
}
```

### Get Alerts (Status Transitions)

```bash
GET /api/v1/alerts?since=2024-02-20T12:00:00Z

curl localhost:8080/api/v1/alerts | jq
```

Response:
```json
{
  "alerts": [
    {
      "processor_id": "processor_a",
      "from_status": "HEALTHY",
      "to_status": "DOWN",
      "timestamp": "2024-02-20T12:15:00Z",
      "reason": "High error/timeout rate (>50%)"
    }
  ],
  "count": 1
}
```

### List Processors

```bash
GET /api/v1/processors

curl localhost:8080/api/v1/processors | jq
```

## Health Calculation Algorithm

### Rolling Window
- Uses last **50 transactions** OR last **10 minutes**
- Ensures recent performance is weighted appropriately

### Authorization Rate
```
auth_rate = approved / (approved + declined)
```
- Errors/timeouts are **not** counted as declines (they're technical failures)
- Separate `error_rate = errors / total` is calculated

### Status Thresholds

| Status | Authorization Rate | Error Rate |
|--------|-------------------|------------|
| **HEALTHY** | >= 65% | < 30% |
| **DEGRADED** | 30-65% | 30-50% |
| **DOWN** | < 30% | > 50% |

Minimum 10 transactions required before changing status (prevents fluctuations).

## Routing Algorithm

1. **Filter** processors by `payment_method` + `country`
2. **Score** each processor:
   ```
   score = auth_rate * 100
   if status == DOWN:     score = 0
   if status == DEGRADED: score *= 0.5
   if transactions > 30:  score += 5  // confidence bonus
   ```
3. **Rank** by score descending
4. **Recommend** top processor (unless all are DOWN)

## Mock Processors

| ID | Name | Countries | Payment Methods |
|----|------|-----------|-----------------|
| processor_a | GlobalPay_BR | BR | PIX, CARD |
| processor_b | PayLatam | BR, MX, CO | CARD |
| processor_c | PixMaster | BR | PIX |
| processor_d | MexPago | MX | CARD, OXXO |
| processor_e | ColombiaPS | CO | PSE, CARD |

## Demo: Outage Scenario

```bash
# Terminal 1: Start server
go run cmd/server/main.go

# Terminal 2: Run demo
./scripts/demo.sh
```

The demo simulates:
1. **Normal operation** - All processors healthy (75% auth rate)
2. **Outage** - processor_a gets 90% errors → transitions to DOWN
3. **Routing shift** - System recommends processor_c for PIX/BR
4. **Recovery** - processor_a gradually recovers
5. **Back to normal** - All processors healthy again

## Project Structure

```
├── cmd/server/main.go       # Server entry point
├── internal/
│   ├── domain/models.go     # Domain models
│   ├── health/calculator.go # Health monitoring logic
│   ├── routing/engine.go    # Routing decision engine
│   └── api/handlers.go      # HTTP handlers
├── scripts/
│   ├── generate_data.go     # Test data generator
│   └── demo.sh              # Demo script
└── dev/                     # Development workflow files
```

## Design Decisions

1. **In-memory storage** - Simplicity over persistence for this challenge
2. **Rolling window** - Balances responsiveness with stability
3. **Separate error rate** - Technical failures vs business declines
4. **Minimum transactions** - Prevents status fluctuation on low volume
5. **Go stdlib** - No external dependencies, uses Go 1.22+ routing

## What I'd Improve With More Time

- [ ] Persistent storage (Redis for real-time, PostgreSQL for history)
- [ ] Circuit breaker pattern with automatic recovery probes
- [ ] Geographic health tracking (per country/region)
- [ ] Anomaly detection (sudden drops even above threshold)
- [ ] WebSocket for real-time dashboard updates
- [ ] Prometheus metrics for monitoring
- [ ] Rate limiting and authentication
