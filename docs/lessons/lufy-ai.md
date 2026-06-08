# Lessons learned de lufy-ai

Registro vivo de aprendizajes operativos que no ameritan ADR completa, pero sí deben reducir redescubrimiento en próximas propuestas, reviews y releases.

## Entradas seed

### Versiones publicadas y documentación copiable deben tener una fuente canónica

- **Contexto:** README y guías de instalación pueden quedar en versiones distintas después de releases automáticas.
- **Aprendizaje:** los comandos copiables de instalación deben validarse contra `RELEASE_VERSION` en CI/local.
- **Aplicación:** `scripts/check-doc-release-version.sh` revisa README, `docs/installation.md` y README de la CLI.

### Validación local debe reproducir el rango real de PR

- **Contexto:** `git diff --check` local puede omitir whitespace introducido en commits ya existentes de la rama.
- **Aprendizaje:** para cambios destinados a PR contra `develop`, el gate debe usar `origin/develop...HEAD` cuando la rama ya tiene commits y `origin/develop` cuando hay cambios pendientes.
- **Aplicación:** `scripts/validate.sh` centraliza el whitespace PR-aware.

### `project_profile` estructural es aceptación, no preferencia

- **Contexto:** una feature puede compilar pero quedar mal ubicada respecto a carpetas, capas o arquitectura seleccionada.
- **Aprendizaje:** instrucciones de estructura del usuario y `project_profile.surfaces[*]` deben convertirse en checklist de aceptación.
- **Aplicación:** router, implementer, validator y reviewer preservan `structural_acceptance` y bloquean readiness si falta estructura obligatoria.

### Engram debe ser opcional y evidenciable

- **Contexto:** la memoria puede no estar instalada o no exponerse en una sesión local.
- **Aprendizaje:** no se debe bloquear trabajo normal ni afirmar trazabilidad de memoria sin herramienta disponible.
- **Aplicación:** agentes tratan Engram como índice opcional y reportan `not_available` cuando no existe evidencia.
