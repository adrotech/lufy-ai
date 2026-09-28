# Roadmap de `lufy-ai`

Este roadmap contiene líneas de evolución posteriores a `v0.6.24`. No es un backlog ejecutable ni implica compromiso de fecha. Toda iniciativa sustantiva debe entrar por routing T1/T2/T3 y adquirir su propia spec, issue, criterios observables y evidencia.

## Principios

- Lufy sigue siendo un harness neutral; tools y metodologías son adapters reemplazables.
- La seguridad y la autoridad humana tienen precedencia sobre autonomía y throughput.
- Las capacidades nuevas deben ser locales-first, explicables, degradables y privadas por default.
- Los estados derivados —Context Graph, proyecciones, reportes y registry— deben poder reconstruirse desde fuentes canónicas.
- Una capacidad sólo pasa al README cuando está instalada por la CLI, validada y publicada.

## Líneas prioritarias

### 1. Paridad de experiencia entre adapters

- reducir diferencias de UX entre OpenCode y Codex sin duplicar el dominio;
- explorar Observatory y reporting para Codex mediante contratos portables;
- evaluar adapters escribibles adicionales sólo cuando exista una superficie segura y verificable.

### 2. Contexto y trazabilidad más profundos

- evolucionar Context Graph hacia señales semánticas opcionales sin perder el baseline determinístico;
- mejorar proyecciones de cobertura scenario-task-test y explicabilidad de impacto;
- conectar observabilidad humana con Run Ledger sin convertir una proyección en autoridad.

### 3. Adaptive routing gobernado

- medir utilidad real de `shadow` antes de ampliar `advisory`;
- mejorar simulación, diagnóstico de capacidad y visualización de leases/budget;
- conservar roles, permisos y gates como fronteras no negociables;
- no habilitar autonomía opaca ni promoción automática de decisiones.

### 4. Paquetes reutilizables

- templates instalables por stack/capability sólo cuando tengan assets, validación y ownership claros;
- subagentes de dominio únicamente cuando reduzcan carga cognitiva sin fragmentar el sistema;
- distribución y descubrimiento de skills externos con consentimiento y provenance verificable.

### 5. Distribución y supply chain

- evaluar canales como Homebrew o Scoop después de estabilizar mantenimiento multiplataforma;
- estudiar verificación cosign integrada en `upgrade` sin degradar la ruta de checksum;
- mantener E2E sobre artifacts publicados y evitar canales que bifurquen la fuente de verdad.

### 6. Operación y métricas

- derivar métricas de ciclo, espera, rework y revisión desde eventos content-free;
- validar que cualquier optimización reduzca tiempo total sin aumentar defectos o carga humana;
- hacer explícitos los casos `not_available`, `stale`, `blocked` y `degraded`.

## Fuera de alcance hasta nueva decisión

- autonomía de delivery sin autorización humana;
- ingestión de transcripts o contenido privado en Run Ledger;
- reemplazo de OpenTelemetry por el ledger local;
- metodologías o adapters que dupliquen roles y contracts del core;
- promesas de templates, marketplace o integraciones antes de existir como assets verificables.

## Mantenimiento

El estado publicado vive en [`status.md`](status.md); el flujo actual vive en [`harness-workflow.md`](harness-workflow.md); decisiones durables viven en ADRs/specs. Al entregar una línea de este roadmap, se elimina de aquí o se reformula como siguiente hipótesis: no se conserva como checklist completada.
