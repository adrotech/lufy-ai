## Why

La distribución pública ya genera binarios versionados, checksums y tags automáticos, pero la cadena de release aún depende de acciones no pineadas, artifacts sin firma/provenance/SBOM y una política de versionado automática demasiado rígida. Antes de ampliar canales de instalación conviene hacer que cada release estable sea auditable, verificable y menos vulnerable a manipulación de workflow o metadata.

## What Changes

- Firmar artifacts y el archivo de checksums con `cosign` keyless usando OIDC de GitHub Actions.
- Generar provenance SLSA para los artifacts publicados y adjuntarla al release.
- Generar SBOM por release y adjuntarla como artifact verificable.
- Pinear acciones de GitHub Actions a commit SHA en workflows de release/auto-tag/CI relevante.
- Reducir permisos de `auto-release-tag.yml` al mínimo necesario y documentar por qué requiere cada permiso.
- Endurecer el auto-tag con labels de bump (`release:patch`, `release:minor`, `release:major`, `release:skip`), retry/backoff ante races y sanitización de títulos de PR en anotaciones.
- Generar release notes consistentes para tags/releases automáticos sin inventar contenido.

## Capabilities

### New Capabilities
- `release-supply-chain-security`: firma keyless, provenance SLSA, SBOM, pinning de actions y permisos mínimos para la cadena de release.
- `release-version-governance`: política de auto-tag basada en labels, skip explícito, retry/backoff, sanitización de metadata y release notes.

### Modified Capabilities
- `versioned-binary-distribution`: los artifacts versionados pasan de checksums SHA-256 simples a artifacts verificables con firma, provenance y SBOM.
- `go-cli-install-ci`: los workflows de CI/release deben ejecutar con actions pineadas y evidencia de validación supply-chain cuando el alcance toque release.

## Impact

- `.github/workflows/release.yml`: permisos, actions pineadas, firma, provenance, SBOM y release notes/artifacts adicionales.
- `.github/workflows/auto-release-tag.yml`: permisos mínimos, selección de bump por labels, skip, retry/backoff, sanitización de anotaciones y dispatch seguro.
- `.github/workflows/go-cli-install.yml`: pinning de actions y coherencia de validación si aplica al alcance.
- `tools/lufy-cli-go/scripts/build-release-artifacts.sh` y `tools/lufy-cli-go/scripts/smoke-release-artifacts.sh`: posible inclusión/verificación de SBOM, signatures y provenance.
- Documentación operativa de release y backlog/OpenSpec asociado.
- Dependencias/tooling de CI: `cosign`/SLSA/SBOM generator según la opción mínima adoptada.
