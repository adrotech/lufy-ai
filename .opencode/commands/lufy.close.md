---
description: Finaliza un cambio activo del workflow LUFY o respaldado por OpenSpec con validación, sync de specs, delivery, PR cerrado y limpieza segura de rama.
agent: orchestrator
---

Finaliza un cambio activo del workflow LUFY o respaldado por OpenSpec a través de los gates de cierre.

## Comportamiento del comando

- Resolver el nombre del cambio desde el argumento o el contexto activo.
- Usar el skill concreto `lufy.close`.
- Tratar este flujo como cierre/finalización invocado por el usuario, no como autorización automática de delivery.
- Verificar tasks, artifacts, validación, sync de specs, estado de delivery, estado de PR y branch safety antes de archivar.
- Verificar que no queden docs/specs/workflow artifacts sin commit ni commits locales sin push cuando esos cambios sean parte del cierre.
- Si los delta specs no están sincronizados, ejecutar o solicitar `/opsx-sync <change>` antes de archivar.
- Si existe evidencia de PR, verificar que esté cerrado o mergeado antes de limpieza/archive readiness; si no se requiere PR, registrar `not_applicable`.
- Limpiar ramas locales/remotas solo cuando el borrado de rama esté explícitamente autorizado y sea seguro; si no, reportar la acción exacta como `delivery_pending`.
- Usar `/opsx-archive <change>` solo después de que pasen los gates de cierre.
- Preservar contexto del repo: CLI Go vive en `tools/lufy-cli-go`; `scripts/install.sh` es un wrapper estricto sin fallback legacy.
- Spec activa/foco actual: `install-managed-assets-with-hash-idempotency`.

## Ejecución recomendada

1. Usar skill `lufy.close`.
2. Si el cierre queda bloqueado, reportar el menor comando de recuperación siguiente.
3. Si la limpieza de rama requiere comandos Git destructivos, enrutar por `delivery` con autorización explícita.
