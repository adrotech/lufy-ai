# Bloque 1: regresiones offline de Codex en CI

## Why
Las 68 pruebas del evaluador pasaron localmente, pero CI no observa su carpeta ni las ejecuta. Una regresión del piloto puede pasar los checks Go.

## What Changes
- Incluir tools/codex-behavior-pilot/** en filtros PR/push y ejecutar unittest desde scripts/validate.sh.
- Python ausente falla con recovery claro; declarar la dependencia en Ubuntu.
- Documentar ejecución sin modelos y plan aprobado de siete bloques Codex.

## Capabilities
### Modified Capabilities
- codex-behavior-pilot: validación offline automática.

## Aceptación y workload
B1-AC1 WHEN cambia el piloto THEN sus paths disparan CI en PR y push.
B1-AC2 WHEN corre scripts/validate.sh THEN unittest propaga fallos sin invocar modelos.
B1-AC3 WHEN falta Python THEN falla con diagnóstico explícito, sin skip exitoso.

Slice T2 del programa T1 autorizado globalmente para implementación, validación y delivery. Seis paths de workflow, script, documentación y spec/mirror forman una unidad coherente; separar comando y disparador no aporta valor. workflow_limits not_available. Overview HTML not_available.
