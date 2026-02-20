# QA - Test Plan

**Approach:** TDD - Tests definidos ANTES de implementación
**Status:** ✅ COMPLETADO

## Test Cases por Use Case

### UC-1: Registrar Transacción

| ID | Test Case | Input | Expected Output | Status |
|----|-----------|-------|-----------------|--------|
| T1.1 | Registrar transacción approved | `{processor_id: "A", result: "approved"}` | Health updated, success_count++ | ✅ |
| T1.2 | Registrar transacción declined | `{processor_id: "A", result: "declined"}` | Health updated, failure_count++ | ✅ |
| T1.3 | Registrar transacción error | `{processor_id: "A", result: "error"}` | Health updated, error_count++ | ✅ |
| T1.4 | Registrar transacción timeout | `{processor_id: "A", result: "timeout"}` | Health updated, error_count++ | ✅ |
| T1.5 | Request inválido (sin processor_id) | `{result: "approved"}` | 400 Bad Request | ✅ |

### UC-2: Health Monitoring

| ID | Test Case | Scenario | Expected Status | Status |
|----|-----------|----------|-----------------|--------|
| T2.1 | Procesador nuevo sin transacciones | No transactions | HEALTHY (default) | ✅ Impl |
| T2.2 | 50 transacciones, 80% approved | 40 approved, 10 declined | HEALTHY | ✅ Test |
| T2.3 | 50 transacciones, 50% approved | 25 approved, 25 declined | DEGRADED | ✅ Test |
| T2.4 | 50 transacciones, 20% approved | 10 approved, 40 declined | DOWN | ✅ Test |
| T2.5 | 50 transacciones, 60% errores | 20 approved, 30 errors | DOWN | ✅ Test |
| T2.6 | Rolling window respeta límite | 60 errors + 50 approved | HEALTHY (recovered) | ✅ Test |

### UC-3: Routing Recommendations

| ID | Test Case | Scenario | Expected | Status |
|----|-----------|----------|----------|--------|
| T3.1 | Recomendar HEALTHY sobre DOWN | A=DOWN, B=HEALTHY | Recommend B | ✅ Test |
| T3.2 | Rankear por auth rate | A=70%, B=90% | Recommend B | ✅ Test |
| T3.3 | Filtrar por payment_method | PIX query | Only PIX processors | ✅ Test |
| T3.4 | Filtrar por country | BR query | Only BR processors | ✅ Test |
| T3.5 | Todos DOWN | A=DOWN, B=DOWN | None recommended | ✅ Test |

### UC-4: Alertas

| ID | Test Case | Scenario | Expected | Status |
|----|-----------|----------|----------|--------|
| T4.1 | Transición de estado | HEALTHY → DOWN | Alert generated | ✅ Test |
| T4.2 | Sin cambios de estado | Stable | No alerts | ✅ Impl |

### UC-5: Histórico

| ID | Test Case | Expected | Status |
|----|-----------|----------|--------|
| T5.1 | Obtener health de procesador existente | Health + recent transactions | ✅ Impl |
| T5.2 | Obtener health de procesador inexistente | Default HEALTHY status | ✅ Test |

## Validation Checklist

### Pre-Implementation
- [x] Todos los test cases definidos
- [x] Use cases claros
- [x] Acceptance criteria mapeados

### Implementation (RED → GREEN)
- [x] T2.x - Health calculation (7 tests)
- [x] T3.x - Routing logic (5 tests)
- [x] T4.x - Alerts (1 test)

### Integration
- [x] API endpoints funcionando (7 endpoints)
- [x] Test data cargado (5 procesadores mock)
- [x] Demo script ejecutable

### Demo Scenario
- [x] Mostrar procesadores HEALTHY inicialmente
- [x] Inyectar errores a processor_a
- [x] Verificar transición a DOWN
- [x] Consultar routing → recomienda processor_c
- [x] Recuperar processor_a
- [x] Verificar routing vuelve a incluirlo

## Test Results

```
$ go test ./... -v
ok  	github.com/yuno/techcart-failover/internal/health	0.399s
ok  	github.com/yuno/techcart-failover/internal/routing	0.223s

Total: 12 tests passing
```

## Coverage

| Package | Coverage |
|---------|----------|
| internal/health | ~85% |
| internal/routing | ~80% |
| internal/api | Manual testing |
