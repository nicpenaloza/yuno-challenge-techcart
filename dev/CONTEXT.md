# CONTEXT - TechCart Failover Intelligence API

## Challenge Summary

**Cliente:** TechCart - Marketplace de electrónicos en Brasil, México y Colombia
**Problema:** Perdieron $420,000 durante un outage de 4 horas porque su routing es estático
**Solución:** API de routing inteligente con monitoreo de salud en tiempo real

## Use Cases (para TDD)

### UC-1: Registrar Resultado de Transacción
**Actor:** Sistema de pagos de TechCart
**Precondición:** Procesador existe en el sistema
**Flujo:**
1. Sistema envía resultado de transacción (processor_id, result, payment_method, country, amount, timestamp)
2. API registra la transacción
3. API recalcula health del procesador
4. API retorna estado actualizado del procesador

**Variantes:**
- UC-1.1: Resultado = approved → incrementa success count
- UC-1.2: Resultado = declined → incrementa failure count (normal, no afecta health)
- UC-1.3: Resultado = error/timeout → incrementa error count (afecta health negativamente)

### UC-2: Consultar Health de Procesadores
**Actor:** Equipo de operaciones de TechCart
**Flujo:**
1. Usuario consulta GET /health
2. API retorna lista de procesadores con: status, authorization_rate, transaction counts
3. Status puede ser: HEALTHY (>=65%), DEGRADED (30-65%), DOWN (<30% o >50% errores)

### UC-3: Obtener Recomendación de Routing
**Actor:** Sistema de pagos de TechCart
**Precondición:** Al menos un procesador soporta el payment_method + country
**Flujo:**
1. Sistema envía consulta: payment_method, country, amount
2. API filtra procesadores que soportan esa combinación
3. API rankea por: status (HEALTHY > DEGRADED > DOWN) y authorization_rate
4. API retorna lista ordenada con recomendación

**Reglas de negocio:**
- Procesador DOWN nunca es recomendado (recommended=false)
- Si todos están DOWN, retorna lista vacía o warning
- El de mayor authorization_rate entre los HEALTHY es el recomendado

### UC-4: Ver Alertas/Transiciones de Estado
**Actor:** Equipo de operaciones
**Flujo:**
1. Usuario consulta GET /alerts?since=<timestamp>
2. API retorna transiciones de estado desde ese momento
3. Cada alerta incluye: processor_id, from_status, to_status, timestamp, reason

### UC-5: Ver Histórico de Procesador
**Actor:** Equipo de operaciones
**Flujo:**
1. Usuario consulta GET /health/{processorId}
2. API retorna: health actual + últimas N transacciones
3. Permite analizar tendencias

## Decisiones de Diseño Preliminares

### Rolling Window para Health
- **Opción A:** Últimas N transacciones (ej: 50)
- **Opción B:** Ventana de tiempo (ej: últimos 10 minutos)
- **Decisión:** Usar ambos - ventana de tiempo Y máximo N transacciones

### Cálculo de Authorization Rate
- Formula: `approved / (approved + declined)`
- Errores/timeouts NO cuentan como declined - son problemas técnicos
- Error rate se calcula aparte: `errors / total`

### Thresholds de Status
| Status | Authorization Rate | Error Rate |
|--------|-------------------|------------|
| HEALTHY | >= 65% | < 30% |
| DEGRADED | 30-65% OR | 30-50% |
| DOWN | < 30% OR | > 50% |

### Procesadores Mock (Test Data)
| ID | Name | Countries | Payment Methods |
|----|------|-----------|-----------------|
| processor_a | GlobalPay_BR | BR | PIX, CARD |
| processor_b | PayLatam | BR, MX, CO | CARD |
| processor_c | PixMaster | BR | PIX |
| processor_d | MexPago | MX | CARD, OXXO |
| processor_e | ColombiaPS | CO | PSE, CARD |

### Escenario de Outage (Demo)
1. Horas 0-2: Todos HEALTHY (70-80% auth rate)
2. Hora 2-4: processor_a (PIX Brazil) → errores masivos → DOWN
3. Hora 4-5: processor_a recupera gradualmente → DEGRADED → HEALTHY
4. Durante outage: processor_c debe ser recomendado para PIX/BR

## Stack Técnico

- **Lenguaje:** Go 1.21+
- **Framework:** net/http (stdlib)
- **Storage:** In-memory (maps con mutex)
- **API:** REST JSON

## Entregables

1. ✅ Backend service con endpoints REST
2. ✅ Test data con escenario de outage
3. ✅ README con setup, API docs, algoritmos
4. ✅ Demo script mostrando failover
