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

### D3: Authorization Rate sin Errores
**Fecha:** 2024-02-20
**Decisión:** `auth_rate = approved / (approved + declined)`
**Razón:** Errores/timeouts son problemas técnicos, no rechazos de negocio
**Trade-off:** Necesitamos métricas separadas para error_rate

### D4: Thresholds de Status
**Fecha:** 2024-02-20
**Decisión:**
- HEALTHY: auth_rate >= 65% AND error_rate < 30%
- DEGRADED: auth_rate 30-65% OR error_rate 30-50%
- DOWN: auth_rate < 30% OR error_rate > 50%
**Razón:** Basado en el challenge que menciona 65-85% como "healthy"
**Trade-off:** Thresholds fijos vs configurables

### D5: Go net/http stdlib
**Fecha:** 2024-02-20
**Decisión:** Usar net/http en lugar de framework (gin, chi, etc.)
**Razón:** Go 1.22+ tiene buen routing, menos dependencias
**Trade-off:** Menos features pero más simple

## Decisiones Pendientes

- [ ] ¿Mínimo de transacciones antes de cambiar status? (evitar fluctuaciones)
- [ ] ¿Persistir alertas o solo en memoria?

## Notas

_Agregar decisiones durante implementación_
