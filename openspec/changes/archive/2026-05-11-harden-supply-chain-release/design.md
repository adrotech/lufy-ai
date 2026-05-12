## Context

El repositorio ya tiene distribución versionada mediante `.github/workflows/release.yml`, artifacts OS/arch, checksums SHA-256, `lufy-ai version`, smoke de artifacts y bootstrap. También tiene auto-tag al cerrar PRs hacia `main` mediante `.github/workflows/auto-release-tag.yml`.

La brecha actual está en garantías de supply chain y gobernanza de versionado: las acciones usan tags flotantes, los artifacts no están firmados, no hay provenance/SBOM publicados y el auto-tag siempre incrementa patch sin permitir skip/minor/major. Esto impacta la confianza de releases públicas y la trazabilidad antes de sumar canales como package managers.

## Goals / Non-Goals

**Goals:**
- Publicar releases con artifacts, checksums, firmas keyless, provenance SLSA y SBOM verificables.
- Pinear actions por SHA en workflows sensibles y mantener comentarios con el tag/version auditada.
- Reducir permisos de workflows, especialmente auto-tag, al mínimo compatible con crear tags y disparar release.
- Controlar version bump desde labels de PR hacia `main`, incluyendo `release:skip`.
- Evitar races de tags con retry/backoff y no sobrescribir tags existentes.
- Sanitizar metadata humana usada en tag annotations y release notes.
- Mantener validación local/CI honesta sin depender de toolchains inexistentes en la raíz.

**Non-Goals:**
- No agregar Homebrew, Scoop u otros package managers.
- No cambiar el contrato de `scripts/install.sh` ni reintroducir fallback legacy.
- No publicar releases desde `develop` ni cambiar la regla de tags `v*` alcanzables desde `main`.
- No exigir firma local con claves mantenidas por humanos; el alcance apunta a keyless/OIDC en CI.

## Decisions

1. Usar firma keyless con OIDC en CI.
   - Rationale: evita secretos de larga vida y se integra con GitHub Actions para artifacts públicos.
   - Alternativa descartada: claves privadas en secrets, porque aumentan rotación y riesgo operacional.

2. Adjuntar firmas/provenance/SBOM como assets del GitHub Release.
   - Rationale: mantiene GitHub Releases como fuente de verdad de distribución y permite verificación offline posterior.
   - Alternativa descartada: solo guardar evidencia como workflow artifact, porque expira y no acompaña al release estable.

3. Mantener SHA-256 checksums aunque se agreguen firmas.
   - Rationale: el bootstrap actual depende de checksums y la firma complementa integridad/autenticidad sin romper el flujo existente.

4. Pinear actions a commit SHA con comentario de versión humana.
   - Rationale: reduce riesgo de supply-chain por tags mutables y conserva mantenibilidad visual.
   - Alternativa descartada: depender de tags semver (`@v4`, `@v5`) en workflows sensibles.

5. Usar labels `release:patch`, `release:minor`, `release:major`, `release:skip` como fuente de bump.
   - Rationale: permite control explícito sin parsear títulos de PR o commits.
   - Regla: si no hay label de bump, default conservador `patch`; si hay `release:skip`, no se crea tag.
   - Conflicto: múltiples labels de bump deben bloquear con error accionable.

6. Sanitizar texto humano antes de usarlo en tag annotations/release notes.
   - Rationale: PR titles pueden incluir saltos de línea, secuencias raras o contenido que degrade logs/anotaciones.
   - Enfoque: normalizar a una línea, recortar longitud y preservar número de PR/commit como identificadores confiables.

## Risks / Trade-offs

- [Tooling supply-chain agrega complejidad] -> Mitigar con pasos pequeños, smoke local cuando exista y documentación de comandos exactos.
- [Pinning por SHA dificulta upgrades] -> Mitigar con comentarios `# actions/checkout@v4` y tarea explícita para revisar SHAs periódicamente.
- [SLSA/cosign pueden requerir permisos adicionales] -> Mitigar declarando permisos mínimos (`id-token: write`, `contents: write`, `attestations` si aplica) solo en el workflow que los usa.
- [Auto-tag con labels puede bloquear promociones si faltan labels] -> Mantener default `patch` y usar `release:skip` para omitir explícitamente.
- [Retry/backoff puede crear tags duplicados si no re-fetch antes de cada intento] -> Re-fetch tags y verificar remoto antes de cada intento; nunca usar force push.

## Migration Plan

1. Crear specs y tareas de supply-chain/version governance.
2. Actualizar workflows/scripts en una rama feature contra `develop`.
3. Validar con `scripts/validate.sh` y, cuando aplique, ejecutar smoke de release local (`build-release-artifacts.sh` / `smoke-release-artifacts.sh`).
4. Abrir PR a `develop` con evidencia; no crear tags estables desde esta rama.
5. Tras merge y promoción futura `develop` -> `main`, verificar que auto-tag y release generen los nuevos assets.

## Open Questions

- Qué generador de SBOM se prefiere (`syft`, `cyclonedx-gomod`, `go version -m` complementario) según facilidad de CI y formato esperado.
- Si se usará `slsa-framework/slsa-github-generator` o GitHub artifact attestations nativas para provenance, según compatibilidad con release assets y permisos disponibles.
