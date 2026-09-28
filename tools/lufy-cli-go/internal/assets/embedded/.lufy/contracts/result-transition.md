# Result Transition V1

`result-transition/v1` es el protocolo tipado para proponer un cambio de estado sin mezclar control de concurrencia con el envelope `result-contract/v1`.

## Intent

Cada intent declara:

- `expected_version` y `previous_fingerprint` para compare-and-swap;
- `idempotency_key` estable para retry seguro;
- actor y ownership pseudonimizados;
- `lease` opcional con token digest y expiración;
- `join` opcional con children exactos, terminales y evidencia agrupada;
- el siguiente Result Contract completo.

## Decisiones

La evaluación devuelve `accepted`, `duplicate_noop`, `rejected` o `conflict`. `closed` es terminal. Una recuperación desde `blocked` o `escalated` necesita una versión nueva y evidencia observable.

`lufy-ai result transition --stdin --json` evalúa en modo read-only. `--record` exige receipt y correlación Run Ledger durables; si cualquiera no está disponible devuelve `unavailable` sin promover `decision_id`, versión o fingerprint aceptados.

Run Ledger conserva únicamente fingerprints, status, versión y referencias SHA-256 allow-listed. No conserva el contrato, summary, prompts, paths ni outputs.
