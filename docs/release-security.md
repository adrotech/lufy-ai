# Seguridad de releases

Este documento describe la política operativa para releases estables de `lufy-ai`.

## Fuente de verdad

- Los releases estables se publican desde tags `v*` alcanzables desde `origin/main`.
- `develop` es integración diaria; no se publican releases estables desde `develop` sin promoción a `main`.
- GitHub Releases es la fuente de verdad para artifacts binarios, checksums, SBOM, provenance y firmas.

## Artifacts esperados

Cada release estable debe publicar:

- `lufy-ai_<version>_<os>_<arch>.tar.gz` o `.zip` por plataforma soportada.
- `lufy-ai_<version>_checksums.txt` con SHA-256 de los archives.
- `lufy-ai_<version>_sbom.spdx.json` con dependencias Go del CLI.
- `lufy-ai_<version>_provenance.intoto.jsonl` con subjects y metadata de build.
- `*.bundle` generado por `cosign sign-blob` para cada artifact/checksum/SBOM/provenance publicado.

## Verificación por consumidores

Verificación de checksum:

```bash
shasum -a 256 -c "lufy-ai_<version>_checksums.txt"
```

Verificación de firma keyless con `cosign`:

```bash
cosign verify-blob \
  --bundle "<artifact>.bundle" \
  --certificate-identity-regexp "https://github.com/adrotech/lufy-ai/.github/workflows/release.yml@.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "<artifact>"
```

Si `cosign` no está instalado, la firma no fue verificada localmente. En ese caso solo puede afirmarse que el checksum coincide, no que se verificó autenticidad keyless.

## Labels de release

El workflow `auto-release-tag` se ejecuta al mergear PRs hacia `main`.

- `release:patch`: incrementa patch.
- `release:minor`: incrementa minor y resetea patch a `0`.
- `release:major`: incrementa major y resetea minor/patch a `0`.
- `release:skip`: no crea tag y termina exitosamente.
- Sin label de release: usa `patch` como default conservador.

Reglas:

- `release:skip` no puede combinarse con labels de bump.
- Solo puede existir un label de bump por PR.
- El workflow nunca sobrescribe tags existentes ni usa force push.
- Si otro proceso crea el mismo tag durante la ejecución, el workflow re-fetches tags, recalcula y reintenta de forma acotada.

## Actualización de actions pineadas

Los workflows sensibles de release deben usar third-party actions pineadas por SHA. Para actualizar una action:

1. Resolver el SHA del tag upstream auditado, por ejemplo `git ls-remote https://github.com/actions/checkout.git refs/tags/v4`.
2. Actualizar `uses:` al SHA resultante.
3. Mantener el comentario con la versión humana, por ejemplo `# v4`.
4. Ejecutar `scripts/validate.sh` antes de abrir PR.
