# QA - Test Plan

**Approach:** TDD - Tests definidos ANTES de implementación

## Test Cases por Use Case

### UC-1: Registrar Transacción

| ID | Test Case | Input | Expected Output |
|----|-----------|-------|-----------------|
| T1.1 | Registrar transacción approved | `{processor_id: "A", result: "approved"}` | Health updated, success_count++ |
| T1.2 | Registrar transacción declined | `{processor_id: "A", result: "declined"}` | Health updated, failure_count++ |
| T1.3 | Registrar transacción error | `{processor_id: "A", result: "error"}` | Health updated, error_count++ |
| T1.4 | Registrar transacción timeout | `{processor_id: "A", result: "timeout"}` | Health updated, error_count++ |
| T1.5 | Request inválido (sin processor_id) | `{result: "approved"}` | 400 Bad Request |

### UC-2: Health Monitoring

| ID | Test Case | Scenario | Expected Status |
|----|-----------|----------|-----------------|
| T2.1 | Procesador nuevo sin transacciones | No transactions | HEALTHY (default) |
| T2.2 | 50 transacciones, 80% approved | 40 approved, 10 declined | HEALTHY |
| T2.3 | 50 transacciones, 50% approved | 25 approved, 25 declined | DEGRADED |
| T2.4 | 50 transacciones, 20% approved | 10 approved, 40 declined | DOWN |
| T2.5 | 50 transacciones, 60% errores | 20 approved, 30 errors | DOWN |
| T2.6 | 50 transacciones, 40% errores | 30 approved, 20 errors | DEGRADED |

### UC-3: Routing Recommendations

| ID | Test Case | Scenario | Expected |
|----|-----------|----------|----------|
| T3.1 | Un procesador HEALTHY | A=HEALTHY | Recommend A |
| T3.2 | Dos procesadores, uno mejor | A=80%, B=70% | Recommend A, B as fallback |
| T3.3 | Uno HEALTHY, uno DOWN | A=HEALTHY, B=DOWN | Recommend A only |
| T3.4 | Todos DOWN | A=DOWN, B=DOWN | Empty recommendations or warning |
| T3.5 | Filtrar por payment_method | PIX query, A supports PIX, B doesn't | Only A in results |
| T3.6 | Filtrar por country | BR query, A supports BR, B supports MX | Only A in results |

### UC-4: Alertas

| ID | Test Case | Scenario | Expected |
|----|-----------|----------|----------|
| T4.1 | Transición HEALTHY → DEGRADED | Auth rate drops to 50% | Alert generated |
| T4.2 | Transición DEGRADED → DOWN | Auth rate drops to 20% | Alert generated |
| T4.3 | Transición DOWN → HEALTHY | Auth rate recovers to 70% | Alert generated |
| T4.4 | Sin cambios de estado | Stable performance | No alerts |

### UC-5: Histórico

| ID | Test Case | Expected |
|----|-----------|----------|
| T5.1 | Obtener health de procesador existente | Health + recent transactions |
| T5.2 | Obtener health de procesador inexistente | Default HEALTHY status |

## Validation Checklist

### Pre-Implementation
- [ ] Todos los test cases definidos
- [ ] Use cases claros
- [ ] Acceptance criteria mapeados

### Implementation (RED → GREEN)
- [ ] T1.x - Transaction recording
- [ ] T2.x - Health calculation
- [ ] T3.x - Routing logic
- [ ] T4.x - Alerts
- [ ] T5.x - History

### Integration
- [ ] API endpoints funcionando
- [ ] Test data cargado
- [ ] Demo script ejecutable

### Demo Scenario
- [ ] Mostrar procesadores HEALTHY inicialmente
- [ ] Inyectar errores a processor_a
- [ ] Verificar transición a DOWN
- [ ] Consultar routing → debe recomendar processor_c
- [ ] Recuperar processor_a
- [ ] Verificar routing vuelve a incluirlo

## Coverage Target

- Health Calculator: 80%+
- Routing Engine: 80%+
- API Handlers: 70%+
