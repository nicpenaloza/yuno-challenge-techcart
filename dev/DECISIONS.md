# DECISIONS - Technical Decisions Log

## Decisiones Tomadas

### D1: Storage In-Memory
**Fecha:** 2024-02-20
**Decisión:** Usar maps con sync.RWMutex en lugar de base de datos
**Razón:** Challenge de 1 hora, simplicidad > escalabilidad
**Trade-off:** No persiste entre reinicios, pero cumple los requisitos

### D2: Rolling Window Híbrido
**Fecha:** 2024-02-20
**Decisión:** Usar ventana de últimas 50 transacciones O últimos 10 minutos
**Razón:**
- 50 transacciones da suficiente muestra estadística
- 10 minutos evita que datos muy viejos afecten
**Trade-off:** Más complejo que solo contar, pero más preciso
**Implementado en:** `internal/health/calculator.go:13-14`

### D3: Authorization Rate sin Errores
**Fecha:** 2024-02-20
**Decisión:** `auth_rate = approved / (approved + declined)`
**Razón:** Errores/timeouts son problemas técnicos, no rechazos de negocio
**Trade-off:** Necesitamos métricas separadas para error_rate
**Implementado en:** `internal/health/calculator.go:95-100`

### D4: Thresholds de Status
**Fecha:** 2024-02-20
**Decisión:**
- HEALTHY: auth_rate >= 65% AND error_rate < 30%
- DEGRADED: auth_rate 30-65% OR error_rate 30-50%
- DOWN: auth_rate < 30% OR error_rate > 50%
**Razón:** Basado en el challenge que menciona 65-85% como "healthy"
**Trade-off:** Thresholds fijos vs configurables
**Implementado en:** `internal/health/calculator.go:15-19`

### D5: Go net/http stdlib
**Fecha:** 2024-02-20
**Decisión:** Usar net/http en lugar de framework (gin, chi, etc.)
**Razón:** Go 1.22+ tiene buen routing nativo, menos dependencias
**Trade-off:** Menos features pero más simple y portable

### D6: Mínimo de Transacciones
**Fecha:** 2024-02-20
**Decisión:** Requerir mínimo 10 transacciones antes de cambiar status
**Razón:** Evitar fluctuaciones con poco volumen de datos
**Trade-off:** Retrasa detección en procesadores nuevos o de bajo volumen
**Implementado en:** `internal/health/calculator.go:20`

### D7: Routing Score System
**Fecha:** 2024-02-20
**Decisión:** Score = auth_rate * 100, con penalizaciones por status
- DOWN: score = 0 (nunca recomendar)
- DEGRADED: score *= 0.5
- Bonus +5 si tiene >30 transacciones (confianza)
**Razón:** Permite ranking determinístico y flexible
**Trade-off:** No considera costo ni latencia (stretch goal)
**Implementado en:** `internal/routing/engine.go:95-110`

### D8: Alertas como Transiciones de Estado
**Fecha:** 2024-02-20
**Decisión:** Registrar cada cambio de status como HealthTransition
**Razón:** Permite auditoría y debugging de incidentes
**Trade-off:** En memoria, se pierde al reiniciar
**Implementado en:** `internal/health/calculator.go:115-125`

## Decisiones Resueltas

- [x] ¿Mínimo de transacciones antes de cambiar status? → **Sí, 10 transacciones** (D6)
- [x] ¿Persistir alertas o solo en memoria? → **Solo memoria** (aceptable para el challenge)

## Stretch Goals No Implementados (por tiempo)

- Circuit breaker pattern con probes de recuperación
- Routing multi-dimensional (costo, latencia)
- Detección de anomalías (drops súbitos)
- Health por región geográfica
