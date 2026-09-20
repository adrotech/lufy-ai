# Run Ledger causal de agentes

El Run Ledger conserva una historia local, causal y verificable de ejecuciones raíz y subagentes. Su objetivo es responder qué run ocurrió, qué relación tuvo con otros runs, qué estado observable dejó y qué evidencia tipada produjo, sin guardar conversaciones ni contenido sensible.

## Modelo y almacenamiento

La fuente de verdad son eventos inmutables `lufy-run-event/v1` bajo `.lufy/runtime/runs/<run_id>/events/`. Cada evento incluye:

- identidad local: `run_id`, `event_id` y `parent_run_id` opcional;
- causalidad: `caused_by_event_id`, reloj de Lamport y secuencia local;
- clase tipada: `start`, `link`, `checkpoint`, `evidence`, `validation`, `blocked`, `finish` o `warning`;
- referencias por digest a tareas, artefactos y evidencia;
- checkpoint y métricas opcionales con disponibilidad explícita.

Receipts de idempotencia, bindings de adapters y proyecciones son derivados reconstruibles. Los eventos son append-only: `run verify --repair` puede regenerar una proyección, pero nunca reescribe eventos fuente.

`.lufy/runtime/` es estado local ignorado por Git, no es un asset gestionado y no se exporta por red.

## CLI

Registrar un evento tipado desde stdin:

```bash
lufy-ai run record --idempotency-key adapter:event:stable --json
```

Registrar un checkpoint:

```bash
lufy-ai run checkpoint \
  --run run-local \
  --status validated \
  --gate validation \
  --next-owner reviewer \
  --idempotency-key run-local:validated \
  --json
```

Consultar y verificar:

```bash
lufy-ai run status --run run-local --json
lufy-ai run summary --run run-local --json
lufy-ai run verify --run run-local --json
lufy-ai run verify --run run-local --repair --json
```

Previsualizar y aplicar retención:

```bash
lufy-ai run prune --dry-run --json
lufy-ai run prune --yes --json
```

`prune` exige elegir explícitamente preview o ejecución. Nunca elimina un árbol causal que contenga runs activos.

## Idempotencia y concurrencia

Cada productor aporta una clave estable. El primer ingreso devuelve `recorded`; un retry con el mismo fingerprint devuelve `duplicate_noop`; reutilizar la clave con metadata diferente devuelve `conflict` sin sobrescribir el evento anterior.

El store serializa escritores por run con locks portables y lease acotado. Reloj y secuencia se asignan dentro de la sección crítica. Runs diferentes pueden escribirse en paralelo.

## Privacidad

El schema usa allow-list y no contiene un body libre. Se rechazan prompts, respuestas, mensajes, transcripts, argumentos de tools, diffs, contenidos de archivos, secretos y metadata arbitraria. Paths e identidades externas se representan mediante SHA-256 con provenance local; los diagnósticos no reflejan valores rechazados.

La integración Codex consume solo `session_id`, `turn_id`, `agent_id`, `agent_type` y el nombre del lifecycle event. No abre `transcript_path` ni persiste `last_assistant_message`.

## Lifecycle Codex y degradación

Codex mapea `SessionStart`, `SubagentStart`, `SubagentStop`, `Stop` y `SessionEnd` a runs y eventos locales. Los hooks son best-effort: un fallo de configuración, permisos, lock o filesystem produce un warning compacto, mantiene `continue=true` y no avanza gates SDD, validación, delivery ni archive.

Otros adapters usan la frontera portable descrita en `.lufy/contracts/run-ledger-producer.md`; su instrumentación automática queda fuera de esta fase.

## Configuración y límites

Los defaults son conservadores:

- habilitado: `true`;
- raíz: `.lufy/runtime`;
- edad máxima terminal: 30 días;
- máximo de runs terminales: 500;
- máximo total terminal: 64 MiB;
- payload JSON público: 64 KiB;
- referencias por lista: 32;
- lock timeout por defecto: 2 segundos.

La raíz configurada debe permanecer dentro de `.lufy/runtime`. Configuración ausente no significa retención ilimitada.

## Recovery

1. Ejecutar `lufy-ai run verify --run <id> --json` para separar integridad de fuente y freshness de proyección.
2. Si la fuente es healthy y solo falta o está stale el summary, repetir con `--repair`.
3. Si hay eventos, receipts, locks o causalidad corruptos, conservar `.lufy/runtime` y revisar los códigos antes de cualquier limpieza.
4. Usar `run prune --dry-run` antes de toda eliminación explícita.

Deshabilitar productores automáticos no borra historia. La remoción manual de `.lufy/runtime/` es una operación separada y destructiva.

## Relación con OpenTelemetry y Observatory

El ledger es una fuente local de dominio Lufy, no un reemplazo de OpenTelemetry ni un backend de tracing distribuido. Su envelope causal puede alimentar una futura exportación explícita, pero actualmente no realiza networking.

Agent Observatory sigue siendo una proyección TUI efímera orientada a OpenCode. Puede consumir resúmenes del ledger en una etapa futura; no es autoridad sobre eventos, receipts ni gates.

## Compatibilidad cross-platform

El store evita `flock` y usa primitives de filesystem portables. La CI existente ejecuta `go test ./...` y build en Ubuntu, macOS y Windows; localmente `scripts/validate.sh` cubre tests, coverage, build, assets y contratos.
