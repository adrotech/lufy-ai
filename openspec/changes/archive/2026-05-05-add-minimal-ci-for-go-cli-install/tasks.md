## 1. Workflow CI

- [x] 1.1 Crear `.github/workflows/go-cli-install.yml` con triggers para pull requests y pushes relevantes.
- [x] 1.2 Configurar `actions/checkout` y `actions/setup-go` usando `tools/lufy-cli-go/go.mod` como fuente de versión Go.
- [x] 1.3 Ejecutar `go test ./...` desde `tools/lufy-cli-go/`.
- [x] 1.4 Ejecutar `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/`.
- [x] 1.5 Construir `tools/lufy-cli-go/bin/lufy-ai` dentro del job para validar el wrapper Bash.

## 2. Smoke reproducible

- [x] 2.1 Añadir un smoke local versionado para la CLI Go si mantiene el YAML simple y reutilizable.
- [x] 2.2 Validar `install --dry-run --yes --no-engram` contra un temp dir y confirmar que no escribe assets.
- [x] 2.3 Validar install real + `verify --no-engram` contra un temp dir.
- [x] 2.4 Validar idempotencia ejecutando install dos veces y comparando estado/asset clave.
- [x] 2.5 Validar `backup`, `restore --dry-run` y restore real con `--yes` contra un temp dir.
- [x] 2.6 Validar que `install` y `restore` fallan de forma accionable sin `--yes` cuando habría mutaciones reales.

## 3. Wrapper y checks estáticos

- [x] 3.1 Ejecutar `scripts/install.sh <temp> --dry-run --yes --no-engram` usando el binario Go local construido y confirmar ausencia de mutaciones.
- [x] 3.2 Ejecutar `scripts/install.sh --target <temp> --yes --no-engram` y luego `tools/lufy-cli-go/bin/lufy-ai verify --target <temp> --no-engram`.
- [x] 3.3 Ejecutar `openspec list --json` como sanity check del estado OpenSpec.
- [x] 3.4 Ejecutar `git diff --check` para detectar whitespace issues.
- [x] 3.5 Evaluar `shellcheck scripts/install.sh`; si no se incorpora, documentar explícitamente la razón y no inventar disponibilidad local.

## 4. Documentación

- [x] 4.1 Documentar en `tools/lufy-cli-go/README.md` o README equivalente cómo correr localmente tests, build y smoke.
- [x] 4.2 Documentar que la CI no requiere Engram y usa `--no-engram`.
- [x] 4.3 Documentar que no hay comandos Node/TS globales de raíz para este gate.

## 5. Validación final

- [x] 5.1 Ejecutar `go test ./...` desde `tools/lufy-cli-go/`.
- [x] 5.2 Ejecutar `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/`.
- [x] 5.3 Ejecutar el smoke local completo si se agregó script, o reproducir los comandos equivalentes del workflow.
- [x] 5.4 Ejecutar validación del wrapper `scripts/install.sh` contra temp dir.
- [x] 5.5 Ejecutar `openspec status --change add-minimal-ci-for-go-cli-install --json` y confirmar que el change queda listo para apply/verify.
- [x] 5.6 Ejecutar `git diff --check` y reportar resultados reales.
