# PR readiness — make-result-contract-executable

## Título propuesto

`feat(result-contract): ejecutar validación y transiciones causales`

## Body propuesto

### Resumen

- convierte `result-contract/v1` en un dominio Go estricto, bounded y canonicalizable;
- agrega `result-transition/v1` con tabla explícita, CAS, ownership, leases, joins e idempotencia;
- expone `lufy-ai result validate|normalize|transition` con salida humana/JSON y exit codes estables;
- reemplaza detección por substring en lifecycle y correlaciona decisiones con Run Ledger sin persistir contenido sensible;
- sincroniza tres specs activas y mantiene contratos/assets root y embedded en paridad.

### Trazabilidad

- Issue: `#220`.
- Change Lufy SDD: `make-result-contract-executable`.
- Specs activas:
  - `executable-result-contract`;
  - `result-transition-validation`;
  - `result-contract-ledger-correlation`.
- No inicia `add-bounded-loop-engine`; esa etapa depende del merge y cierre de este PR.

### Review slices

1. **Schema y canonicalización**
   - `internal/resultcontract/{model,decode,validate,canonical,scalar,diagnostic}.go`
   - fixtures, property/fuzz y artifacts SDD.
   - foco: compatibilidad v1, límites, determinismo y diagnósticos sanitizados.

2. **Policies y protocolo distribuido**
   - `claim.go`, `state.go`, `transition.go`, `ownership.go`, `receipt.go`, `join.go`.
   - foco: matriz de roles/evidencia, 81 edges, fencing, leases, retry atómico y causalidad de joins.

3. **CLI y persistencia explícita**
   - `internal/cli/app_result.go`, `result_receipt_store.go`, exit codes y command palette.
   - foco: read-only por defecto, normalización legacy estrecha y `--record` fail-closed.

4. **Lifecycle, ledger y assets**
   - `internal/codexlifecycle`, `internal/resultcontractledger`, contratos/templates y embedded.
   - foco: parse real, `continue=true`, metadata content-free, privacidad y paridad gestionada.

### Validación

- `go test -timeout 300s ./... -count=1` — passed.
- `go build ./...` — passed.
- race focalizado — passed.
- fuzz bounded de tres ingress — passed.
- `scripts/validate.sh` — passed; coverage global 80,1%.
- cross-compilación Windows de paquetes afectados — passed.
- strict Lufy SDD validate — passed.
- Lufy SDD sync — synced, 14 requirements en tres specs activas.

### Riesgos y recuperación

- Receipt y evento Run Ledger usan dos escrituras locales; un retry `duplicate_noop` completa una correlación interrumpida.
- La validación estructural no prueba la veracidad material de la evidencia; validator/delivery siguen siendo autoridades.
- `shellcheck` no estaba disponible localmente; CI debe ejecutar su gate habitual.
- Windows se cross-compiló, pero la ejecución nativa queda para CI.

### Delivery pendiente

- Reejecutar `lufy-ai pr guard` después de stage/commit: el gate actual aprobó los 16 archivos tracked modificados, pero los archivos nuevos aún están untracked.
- Requiere autorización explícita para commit, push y creación del PR contra `develop`.
- No cerrar ni archivar hasta checks remotos exitosos, merge, sync confirmado y cleanup seguro.

`Refs #220`
