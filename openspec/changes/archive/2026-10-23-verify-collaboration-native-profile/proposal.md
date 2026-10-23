# Perfil nativo collaboration verificable

## Why

Los controles del 2026-10-23 reproducen ausencia de `agent_type` al ignorar la configuración de confianza del fixture. Registrar explícitamente `agents.reviewer.config_file` y su descripción en la invocación vuelve a anunciar reviewer sin cargar configuración de usuario. La API observada ofrece `followup_task`, `list_agents` y `wait_agent`, sin `close_agent`; el protocolo legacy no debe presentarse como exitoso para esa API.

## What Changes

- Añadir `--transport collaboration` opt-in; conservar `legacy` por defecto y su aceptación íntegra.
- Registrar sólo el reviewer copiado al fixture mediante overrides de invocación, conservando `--ignore-user-config`, sandbox read-only y aprobación never; no copiar/enlazar credenciales ni cambiar confianza global.
- Solicitar spawn reviewer, inspección estructurada del resultado vacío, una recuperación al mismo target y otra inspección del contrato esperado. Usar callbacks de `list_agents`, no texto del modelo, como evidencia del payload.
- Separar transporte, rol y lifecycle observados; cierre `not_available` mantiene la cadena completa inconclusa aun cuando el transporte pase.
- Ejecutar pruebas offline adversariales y un único ensayo live acotado tras validación; publicar también resultados incompletos.

## Impact

Capability `codex-native-regression`; runner/normalizador/evaluador, tests, README, reporte y spec/mirror. Sin cambios al producto CLI Go, modelos, credenciales, permisos o configuración global.

## Workflow

T1 OpenSpec completo: nuevo perfil de evidencia con riesgo de aislamiento. Dirección aprobada explícitamente por el usuario; no requiere otra aprobación por recuento de archivos. `workflow_limits` no disponible; graph del worktree no disponible, fallback dirigido a rutas ya analizadas. Dos review slices secuenciales (registro/normalización y evaluación/evidencia), una entrega coherente porque comparten el mismo protocolo. Sin paralelismo; validación final agrupada. HTML overview: not_available.
