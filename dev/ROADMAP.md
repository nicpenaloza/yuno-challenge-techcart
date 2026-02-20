# ROADMAP - TechCart Failover Intelligence API

**Challenge:** Build intelligent payment routing with health monitoring
**Tiempo:** 1 hora | **Inicio:** 11:47 COL | **Deadline:** 12:47 COL
**Approach:** TDD (Test-Driven Development)

---

## Phase 1: Setup & Domain (5 min) ✅
> Estructura del proyecto y modelos de dominio

- [x] 1.1 Inicializar proyecto Go (go.mod)
- [x] 1.2 Crear estructura de carpetas (cmd/, internal/)
- [x] 1.3 Definir domain models:
  - Transaction (id, processor_id, result, payment_method, country, amount, timestamp)
  - Processor (id, name, countries, payment_methods)
  - ProcessorHealth (processor_id, status, auth_rate, counts, last_updated)
  - HealthStatus enum (HEALTHY, DEGRADED, DOWN)
  - RoutingRecommendation (rankings, payment_method, country)

---

## Phase 2: Core - Health Calculator (15 min) ⬜ [25 pts]
> Lógica de monitoreo de salud - TDD

### 2.1 Tests RED
- [ ] 2.1.1 Test: Procesador nuevo → HEALTHY por defecto
- [ ] 2.1.2 Test: 80% approved → HEALTHY
- [ ] 2.1.3 Test: 50% approved → DEGRADED
- [ ] 2.1.4 Test: 20% approved → DOWN
- [ ] 2.1.5 Test: >50% errores → DOWN
- [ ] 2.1.6 Test: Rolling window respeta límite de 50

### 2.2 Implementation GREEN
- [ ] 2.2.1 Implementar Calculator struct con storage
- [ ] 2.2.2 Implementar RecordTransaction()
- [ ] 2.2.3 Implementar calculateHealth() con rolling window
- [ ] 2.2.4 Implementar determineStatus() con thresholds
- [ ] 2.2.5 Implementar GetHealth() y GetAllHealth()

### 2.3 Refactor
- [ ] 2.3.1 Extraer constantes (thresholds, window size)
- [ ] 2.3.2 Agregar mutex para concurrencia

---

## Phase 3: Core - Routing Engine (10 min) ⬜ [25 pts]
> Lógica de routing inteligente - TDD

### 3.1 Tests RED
- [ ] 3.1.1 Test: Recomendar procesador HEALTHY sobre DOWN
- [ ] 3.1.2 Test: Rankear por authorization_rate
- [ ] 3.1.3 Test: Filtrar por payment_method
- [ ] 3.1.4 Test: Filtrar por country
- [ ] 3.1.5 Test: Todos DOWN → lista vacía o warning

### 3.2 Implementation GREEN
- [ ] 3.2.1 Implementar Engine struct
- [ ] 3.2.2 Implementar RegisterProcessor()
- [ ] 3.2.3 Implementar findCandidates() - filtro por method/country
- [ ] 3.2.4 Implementar rankProcessors() - score por health
- [ ] 3.2.5 Implementar Recommend() - orquestador

---

## Phase 4: API REST (10 min) ⬜ [15 pts]
> Endpoints HTTP

- [ ] 4.1 Crear Handler struct con dependencias
- [ ] 4.2 POST /api/v1/transactions - Registrar transacción
- [ ] 4.3 GET /api/v1/health - Estado de todos los procesadores
- [ ] 4.4 GET /api/v1/health/{processorId} - Estado + histórico
- [ ] 4.5 POST /api/v1/routing/recommend - Obtener recomendación
- [ ] 4.6 GET /api/v1/alerts - Transiciones de estado
- [ ] 4.7 Crear main.go con server setup

---

## Phase 5: Test Data & Demo (10 min) ⬜ [10 pts]
> Datos de prueba y demostración

- [ ] 5.1 Crear procesadores mock (5 procesadores)
- [ ] 5.2 Generar 1000+ transacciones con patrones:
  - Normal: 70-80% auth rate
  - Outage: processor_a con 90% errores por 2 horas
  - Degradación: processor_b bajando gradualmente
- [ ] 5.3 Crear script de demo (demo.sh o demo.go)
- [ ] 5.4 Probar escenario completo de failover

---

## Phase 6: Documentation (5 min) ⬜ [5 pts]
> README y documentación

- [ ] 6.1 README con:
  - Setup instructions
  - API documentation (endpoints, examples)
  - Algoritmos explicados
  - Demo instructions
- [ ] 6.2 Comentarios en código crítico

---

## Progress Tracker

| Phase | Tasks | Done | Progress |
|-------|-------|------|----------|
| 1. Setup | 3 | 3 | ✅ 100% |
| 2. Health | 11 | 0 | ⬜ 0% |
| 3. Routing | 10 | 0 | ⬜ 0% |
| 4. API | 7 | 0 | ⬜ 0% |
| 5. Data | 4 | 0 | ⬜ 0% |
| 6. Docs | 2 | 0 | ⬜ 0% |
| **Total** | **37** | **3** | **8%** |

---

## Checkpoints

- [ ] ⚠️ CHECKPOINT 1: Core Logic funciona (Phase 2+3) → 50 pts secured
- [ ] ⚠️ CHECKPOINT 2: API funciona (Phase 4) → 65 pts secured
- [ ] ⚠️ CHECKPOINT 3: Demo funciona (Phase 5) → 75 pts secured
- [ ] ⚠️ CHECKPOINT 4: Documentación (Phase 6) → 80+ pts secured

---

## Notes

- Priorizar funcionalidad sobre perfección
- Si el tiempo aprieta, saltar tests y ir directo a GREEN
- README mínimo viable es suficiente
