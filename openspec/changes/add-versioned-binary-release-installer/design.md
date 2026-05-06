## Context

`lufy-ai` ya cuenta con una CLI Go en `tools/lufy-cli-go` y un wrapper `scripts/install.sh` que delega estrictamente en `lufy-ai install`. El flujo público actual todavía depende de clonar el repositorio y construir el binario localmente, porque el wrapper solo usa un binario local o uno ya presente en `PATH` y no descarga nada remoto.

La oportunidad es pasar de “clone + build” a una distribución versionada, verificable y compatible con instalación remota. La inspiración externa incluye canales comunes como `curl | bash`, Homebrew, Scoop, `go install` y GitHub Releases con checksums, pero el diseño debe adaptarse al rol real de `lufy-ai`: instalar assets OpenCode/OpenSpec gestionados de forma segura e idempotente en repositorios destino.

El punto crítico para eliminar el clone no es solo publicar el binario: la CLI actual instala assets que viven en el checkout fuente. Para una instalación standalone real, el binario distribuido deberá llevar esos assets embebidos o descargarlos desde un bundle versionado que también esté cubierto por integridad.

## Goals / Non-Goals

**Goals:**

- Definir fases implementables para distribuir `lufy-ai` sin exigir clone del repositorio.
- Publicar artifacts de release versionados para macOS, Linux y Windows cuando el runtime se implemente.
- Verificar integridad con checksums SHA-256 antes de instalar o ejecutar binarios descargados.
- Añadir `lufy-ai version` con versión, commit, fecha de build y plataforma.
- Diseñar un bootstrap remoto que soporte version pinning, detección OS/arch y alternativas inspeccionables a `curl | bash`.
- Definir una estrategia incremental para assets standalone: primero binario + checksums, luego bootstrap, después assets embebidos o bundle versionado.
- Posponer README/getting-started/CLI docs finales hasta que las capacidades existan realmente.

**Non-Goals:**

- Implementar runtime, workflows de release, bootstrap o cambios de README en esta proposal.
- Reintroducir fallback legacy en `scripts/install.sh`.
- Descargar o ejecutar binarios remotos desde el wrapper local sin checksum y autorización explícita.
- Prometer Homebrew, Scoop o `go install` como canales disponibles antes de implementarlos.
- Mantener retrocompatibilidad documental con instrucciones obsoletas de clone/build una vez exista el flujo sin clone.

## Decisions

1. **Fasear distribución antes de instalación standalone total.**
   - Decisión: implementar primero releases binarios versionados con checksums y `lufy-ai version`; después bootstrap remoto; después assets embebidos o bundle.
   - Rationale: reduce riesgo y permite validar integridad/distribución antes de resolver la fuente standalone de assets.
   - Alternativa considerada: publicar inmediatamente un instalador monolítico remoto. Se descarta porque mezclar bootstrap, release y assets en un solo paso dificulta validar seguridad e idempotencia.

2. **Usar SHA-256 como gate obligatorio de descarga.**
   - Decisión: todo artifact descargado por bootstrap deberá compararse contra un archivo de checksums versionado de la misma release antes de instalarse o ejecutarse.
   - Rationale: evita instalar binarios corruptos o alterados y mantiene coherencia con el uso actual de SHA-256 para assets gestionados.
   - Alternativa considerada: confiar solo en TLS/GitHub Releases. Se descarta porque no deja una evidencia de integridad explícita ni permite verificación offline del artifact descargado.

3. **Version pinning explícito.**
   - Decisión: el bootstrap debe aceptar una versión específica (`--version vX.Y.Z` o variable equivalente) y documentar ese modo como opción recomendada para automatización; `latest` puede existir como conveniencia interactiva.
   - Rationale: los usuarios pueden reproducir instalaciones y evitar upgrades implícitos.
   - Alternativa considerada: instalar siempre latest. Se permite solo como atajo documentado con trade-off, no como única ruta.

4. **Bootstrap descargable, pero inspeccionable.**
   - Decisión: documentar `curl | bash` solo junto a una alternativa que descarga el script, permite inspeccionarlo y luego ejecutarlo.
   - Rationale: respeta DX moderna sin ocultar el riesgo operacional de ejecutar scripts remotos.
   - Alternativa considerada: prohibir `curl | bash`. Se descarta porque es un canal común útil si se acompaña de pinning e inspección.

5. **Assets standalone mediante embed como estrategia preferida inicial.**
   - Decisión: recomendar que la primera instalación standalone real use `go:embed` para incluir el catálogo de assets gestionados dentro del binario, manteniendo release bundles como alternativa futura si el tamaño o frecuencia de assets lo exige.
   - Rationale: un binario autocontenido simplifica checksum, reduce fallos por descargas múltiples y evita drift entre binario y assets.
   - Alternativa considerada: bundle zip/tar con binario + assets desde el inicio. Es viable, pero aumenta complejidad del bootstrap y requiere verificar múltiples archivos; puede adoptarse después si `go:embed` se vuelve incómodo.

6. **Canales secundarios después del release base.**
   - Decisión: Homebrew, Scoop y `go install` se diseñan como tareas posteriores al binario versionado + bootstrap, no como prerequisito de la primera release.
   - Rationale: GitHub Releases con checksums es el contrato base que esos canales pueden consumir.

## Risks / Trade-offs

- **Binario publicado sin assets suficientes** → Mitigación: mantener clone/build como estado actual hasta completar embed o bundle; specs exigen que standalone no se declare real hasta resolver assets.
- **`curl | bash` percibido como inseguro** → Mitigación: documentar alternativa inspeccionable, pinning y checksum obligatorio; no auto-run destructivo salvo flag explícito.
- **Matrices OS/arch incompletas** → Mitigación: definir plataformas soportadas explícitamente y fallar con mensaje accionable para combinaciones no soportadas.
- **Drift entre release, checksums y docs** → Mitigación: CI genera artifacts/checksums en un mismo workflow y docs finales se actualizan solo al cierre.
- **Homebrew/Scoop agregan mantenimiento** → Mitigación: dejarlos como fase posterior con aceptación separada.

## Migration Plan

1. Crear `lufy-ai version` y metadata de build inyectada por flags de linker en releases.
2. Añadir workflow de GitHub Actions para build multiplataforma, empaquetado y checksums SHA-256 sin publicar automáticamente desde ramas no autorizadas.
3. Añadir smokes que descarguen/usen artifacts generados localmente por el workflow y validen `version`, `install --dry-run` y `verify` en temporales.
4. Añadir bootstrap remoto con OS/arch detection, version pinning, checksum obligatorio, instalación en PATH y modo dry-run/inspectable.
5. Resolver assets standalone mediante `go:embed` inicialmente; si se elige bundle, verificar checksum del bundle completo y manifest interno antes de instalar.
6. Actualizar README, `docs/getting-started.md` y `tools/lufy-cli-go/README.md` cuando el flujo sin clone esté implementado y validado; borrar instrucciones obsoletas si dejan de aplicar.

## Open Questions

- ¿Qué matriz mínima de plataformas se publicará primero: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`?
- ¿El canal `latest` debe apuntar al último release estable o requerir siempre versión explícita en CI?
- ¿La primera distribución debe incluir Homebrew/Scoop o solo dejar preparada la metadata para fórmulas/manifests posteriores?
