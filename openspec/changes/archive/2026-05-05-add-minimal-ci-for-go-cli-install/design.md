## Context

El repositorio ya contiene la CLI Go real en `tools/lufy-cli-go` y el wrapper `scripts/install.sh` delega estrictamente en ese binario. No hay `.github/workflows/*`, y la raíz no tiene `package.json` ni toolchain Node/TS global que pueda asumirse para validación.

La CI mínima debe proteger el contrato actual del instalador sin ampliar scope funcional: compilar, probar y ejecutar smokes contra directorios temporales para confirmar que `install`, `verify`, `backup`, `restore` y el wrapper siguen funcionando de forma segura.

## Goals / Non-Goals

**Goals:**

- Crear un workflow GitHub Actions pequeño y mantenible para la CLI Go.
- Ejecutar `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go`.
- Construir `tools/lufy-cli-go/bin/lufy-ai` dentro del job para validar `scripts/install.sh` sin depender de binarios preexistentes.
- Ejecutar smoke tests con targets temporales para dry-run sin mutaciones, install real, verify, idempotencia básica, backup y restore.
- Validar que `openspec list --json` no falle y que el repo no tenga whitespace issues con `git diff --check`.
- Documentar los comandos locales equivalentes.

**Non-Goals:**

- No añadir `sync`, `update`, releases, packaging ni publicación de artefactos.
- No introducir Node/TS tooling en la raíz ni asumir `npm test` global.
- No requerir Engram en CI; los smokes deben usar `--no-engram` para mantenerse portables.
- No reemplazar tests unitarios existentes por scripts opacos; el smoke debe complementar `go test`.
- No reintroducir fallback legacy en `scripts/install.sh`.

## Decisions

1. **Workflow único `go-cli-install.yml`**

   Usar un solo workflow evita una matriz prematura y mantiene el gate fácil de auditar. Alternativa considerada: separar test/build/smoke en workflows distintos; se descarta por overhead para la primera CI.

2. **Go version desde `tools/lufy-cli-go/go.mod`**

   `actions/setup-go` SHALL leer `go-version-file: tools/lufy-cli-go/go.mod` para evitar duplicar versiones. Alternativa considerada: hardcodear versión en YAML; se descarta para reducir drift.

3. **Smoke shell inline o script pequeño versionado**

   El smoke puede vivir como script pequeño bajo `tools/lufy-cli-go/scripts/` si el YAML se vuelve difícil de leer. La decisión preferida es encapsular las verificaciones repetibles en un script local para poder correrlo igual en CI y desarrollo.

4. **Targets temporales siempre fuera del repo**

   Los smokes SHALL usar `mktemp -d` o equivalente para no escribir assets instalados en el checkout. Esto protege el worktree y valida instalación real en un target limpio.

5. **Engram omitido en CI**

   Todos los smokes SHALL usar `--no-engram`. La resolución de Engram ya tiene tests unitarios con resolver fake; CI no debe depender de que `engram` exista en el runner.

6. **Wrapper validado con binario local construido**

   El workflow SHALL construir `tools/lufy-cli-go/bin/lufy-ai` antes de invocar `scripts/install.sh`, porque el wrapper debe fallar sin binario y delegar cuando existe. Esto valida el contrato estricto sin instalar nada global.

## Risks / Trade-offs

- **Riesgo: smoke demasiado largo o frágil** → Mantenerlo acotado a flujos críticos y sin red ni dependencias externas.
- **Riesgo: shell portability** → Ejecutar en `ubuntu-latest` con Bash explícito y documentar que es el ambiente canónico inicial.
- **Riesgo: validar menos que una E2E completa** → Cubrir los caminos mínimos: dry-run, install, verify, idempotencia, backup/restore y wrapper; dejar escenarios avanzados para futuras proposals.
- **Riesgo: duplicación entre YAML y docs** → Documentar el script/comandos reales y referenciarlo desde README en vez de copiar lógica larga en varios lugares.
