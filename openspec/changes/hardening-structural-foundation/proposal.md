## Why

El backlog identifica riesgos P1/P2 en la base estructural del instalador: drift entre assets canonicos y embebidos, path safety incompleto para Windows, metadata de instalacion hardcodeada, fingerprint no operativo y escrituras no atomicas. Resolver esta ola primero mejora trazabilidad, reproducibilidad y seguridad antes de ampliar supply chain, CI o UX.

## What Changes

- Definir una fuente de verdad verificable para assets raiz vs assets embebidos, con validacion de paridad para evitar drift.
- Endurecer la normalizacion de paths relativos para bloquear escapes con separadores Windows y rutas ambiguas cuando `EnsureRelativeSafe` se usa directamente.
- Persistir metadata real de version/build en `.lufy-ai/install-state.json` y manifests de backup en vez de `dev` hardcodeado.
- Reemplazar `SourceRootFingerprint` fijo por un fingerprint calculado desde el catalogo ordenado de assets.
- Hacer atomicas las escrituras de archivos gestionados en install, sync y backup usando temp file + rename dentro del mismo directorio.
- Mantener `scripts/install.sh` como wrapper estricto y no introducir fallback legacy.

## Capabilities

### New Capabilities
- `installer-structural-hardening`: Reglas base de hardening para paridad de assets, path safety portable, metadata/fingerprint de install state y escrituras atomicas.

### Modified Capabilities
- `managed-assets-install`: El manifest, el catalogo, backup/restore, verify, install y sync cambian requisitos para metadata real, fingerprint calculado, path safety portable y escrituras atomicas.
- `go-cli-installer`: La CLI Go cambia requisitos de seguridad y validacion para reflejar path safety portable, version metadata en state y escritura atomica.
- `go-cli-install-ci`: La validacion CI/local cambia para incluir paridad de assets, path traversal Windows y atomicidad/fingerprint cuando aplique.

## Impact

- `tools/lufy-cli-go/internal/assets/`
- `tools/lufy-cli-go/internal/platform/`
- `tools/lufy-cli-go/internal/state/`
- `tools/lufy-cli-go/internal/installer/`
- `tools/lufy-cli-go/internal/syncer/`
- `tools/lufy-cli-go/internal/backup/`
- `tools/lufy-cli-go/internal/verify/`
- `tools/lufy-cli-go/internal/version/`
- `.github/workflows/go-cli-install.yml`
- Specs OpenSpec y documentacion asociada al backlog
