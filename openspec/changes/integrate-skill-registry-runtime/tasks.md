## 1. Contrato y primitive

- [x] 1.1 Exponer inspección programática e implementar `Service.Ensure`.
- [x] 1.2 Agregar `lufy-ai skills ensure` y preservar `refresh/status`.
- [x] 1.3 Probar no-op ready, refresh stale/ausente, rutas movidas y Windows.

## 2. Lifecycle y harnesses

- [x] 2.1 Ejecutar ensure best-effort después de install/sync con el tool efectivo.
- [x] 2.2 Integrar el evento `session.created` de OpenCode con un hook best-effort.
- [x] 2.3 Documentar el fallback de primer uso de Codex sin activar hooks no verificados.
- [x] 2.4 Probar ausencia del binario y propagación de warnings no fatales.

## 3. Diagnóstico y documentación

- [x] 3.1 Incorporar el estado read-only a `doctor`.
- [x] 3.2 Incorporar el estado read-only a `verify --deep`.
- [x] 3.3 Actualizar README/status y mantener paridad source/embedded.
- [x] 3.4 Ejecutar validaciones del bloque y registrar cualquier limitación local.
  - Evidencia local: whitespace, checks de release/harness, smoke de hooks, paridad source/embedded y estructura del delta pasaron.
  - Limitación: `scripts/validate.sh` se detuvo en PR guard porque `go` no está instalado; quedan pendientes tests, coverage, vet y build Go en CI o en un entorno con toolchain.
