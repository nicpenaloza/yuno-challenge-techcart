package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080/api/v1"

type Transaction struct {
	ProcessorID   string  `json:"processor_id"`
	Result        string  `json:"result"`
	PaymentMethod string  `json:"payment_method"`
	Country       string  `json:"country"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("🔄 Generating test data for TechCart Failover API")
	fmt.Println("================================================")
	fmt.Println("")

	// Scenario timeline (simulated over "hours")
	// Hour 0-2: Normal operation (all processors healthy)
	// Hour 2-4: processor_a OUTAGE (90% errors)
	// Hour 4-5: processor_a RECOVERY (gradually improving)
	// Hour 5-6: Back to normal

	totalTx := 0

	// Phase 1: Normal operation (400 transactions)
	fmt.Println("📊 Phase 1: Normal operation (all healthy)")
	totalTx += generateNormalTraffic(400)
	printHealth()

	// Phase 2: Outage on processor_a (300 transactions)
	fmt.Println("\n🔴 Phase 2: OUTAGE on processor_a (GlobalPay_BR)")
	totalTx += generateOutageTraffic(300)
	printHealth()
	printRouting("PIX", "BR")

	// Phase 3: Recovery (200 transactions)
	fmt.Println("\n🟡 Phase 3: processor_a RECOVERING")
	totalTx += generateRecoveryTraffic(200)
	printHealth()

	// Phase 4: Back to normal (200 transactions)
	fmt.Println("\n🟢 Phase 4: All processors recovered")
	totalTx += generateNormalTraffic(200)
	printHealth()
	printRouting("PIX", "BR")

	// Print alerts
	fmt.Println("\n🚨 Health Transitions (Alerts):")
	printAlerts()

	fmt.Printf("\n✅ Generated %d transactions total\n", totalTx)
}

func generateNormalTraffic(count int) int {
	processors := []struct {
		id      string
		methods []string
		country string
	}{
		{"processor_a", []string{"PIX", "CARD"}, "BR"},
		{"processor_b", []string{"CARD"}, "BR"},
		{"processor_c", []string{"PIX"}, "BR"},
		{"processor_d", []string{"CARD", "OXXO"}, "MX"},
		{"processor_e", []string{"PSE", "CARD"}, "CO"},
	}

	for i := 0; i < count; i++ {
		p := processors[rand.Intn(len(processors))]
		method := p.methods[rand.Intn(len(p.methods))]
		result := normalResult() // 75% approved, 20% declined, 5% error
		sendTransaction(p.id, result, method, p.country)
	}
	return count
}

func generateOutageTraffic(count int) int {
	// processor_a has 90% errors, others normal
	for i := 0; i < count; i++ {
		// 40% of traffic goes to processor_a (it's primary for PI Brazil)
		if rand.Float32() < 0.4 {
			result := outageResult() // 90% errors
			sendTransaction("processor_a", result, "PIX", "BR")
		} else {
			// Other processors remain healthy
			others := []struct {
				id      string
				method  string
				country string
			}{
				{"processor_b", "CARD", "BR"},
				{"processor_c", "PIX", "BR"},
				{"processor_d", "CARD", "MX"},
				{"processor_e", "PSE", "CO"},
			}
			p := others[rand.Intn(len(others))]
			sendTransaction(p.id, normalResult(), p.method, p.country)
		}
	}
	return count
}

func generateRecoveryTraffic(count int) int {
	// processor_a recovering: 50% errors -> 30% -> 10%
	errorRate := 0.5
	decrement := 0.4 / float64(count)

	for i := 0; i < count; i++ {
		if rand.Float32() < 0.3 {
			var result string
			if rand.Float64() < errorRate {
				result = "error"
			} else {
				result = normalResult()
			}
			sendTransaction("processor_a", result, "PIX", "BR")
			errorRate -= decrement
			if errorRate < 0.05 {
				errorRate = 0.05
			}
		} else {
			// Other processors
			sendTransaction("processor_c", normalResult(), "PIX", "BR")
		}
	}
	return count
}

func normalResult() string {
	r := rand.Float32()
	if r < 0.75 {
		return "approved"
	} else if r < 0.95 {
		return "declined"
	}
	return "error"
}

func outageResult() string {
	if rand.Float32() < 0.90 {
		return "error"
	}
	return "approved"
}

func sendTransaction(processorID, result, method, country string) {
	tx := Transaction{
		ProcessorID:   processorID,
		Result:        result,
		PaymentMethod: method,
		Country:       country,
		Amount:        float64(rand.Intn(10000)) + 100,
		Currency:      currencyFor(country),
	}

	body, _ := json.Marshal(tx)
	resp, err := http.Post(baseURL+"/transactions", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return
	}
	resp.Body.Close()
}

func currencyFor(country string) string {
	switch country {
	case "BR":
		return "BRL"
	case "MX":
		return "MXN"
	case "CO":
		return "COP"
	default:
		return "USD"
	}
}

func printHealth() {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Println("Error fetching health:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	processors, _ := result["processors"].([]interface{})
	fmt.Printf("  Processors: %d\n", len(processors))
	for _, p := range processors {
		proc := p.(map[string]interface{})
		fmt.Printf("    - %s: %s (auth_rate: %.1f%%, txs: %.0f)\n",
			proc["processor_id"],
			proc["status"],
			proc["authorization_rate"].(float64)*100,
			proc["total_transactions"])
	}
}

func printRouting(method, country string) {
	url := fmt.Sprintf("%s/routing/recommend?payment_method=%s&country=%s", baseURL, method, country)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	recs, _ := result["recommendations"].([]interface{})
	fmt.Printf("  Routing for %s/%s:\n", method, country)
	for _, r := range recs {
		rec := r.(map[string]interface{})
		recommended := ""
		if rec["recommended"].(bool) {
			recommended = "✓ RECOMMENDED"
		}
		fmt.Printf("    #%v %s [%s] %.1f%% %s\n",
			rec["rank"],
			rec["processor_id"],
			rec["status"],
			rec["authorization_rate"].(float64)*100,
			recommended)
	}
}

func printAlerts() {
	resp, err := http.Get(baseURL + "/alerts")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	alerts, _ := result["alerts"].([]interface{})
	if len(alerts) == 0 {
		fmt.Println("  No alerts")
		return
	}
	for _, a := range alerts {
		alert := a.(map[string]interface{})
		fmt.Printf("  - %s: %s → %s (%s)\n",
			alert["processor_id"],
			alert["from_status"],
			alert["to_status"],
			alert["reason"])
	}
}
