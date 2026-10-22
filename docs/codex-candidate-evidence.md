# Evidencia y candidato Git en Codex

`candidate capture` liga un Result Contract válido al repositorio, base y HEAD comprometido observados. `candidate verify` vuelve a calcular las identidades y rechaza evidencia obsoleta. Ambos comandos son read-only y no avanzan gates.

```sh
lufy-ai candidate capture --target <checkout-limpio> --base origin/develop --evidence <result-contract.json> --role reviewer --json > /tmp/candidate-evidence.json
lufy-ai candidate verify --target <checkout-limpio> --base origin/develop --evidence <result-contract.json> --role reviewer --binding /tmp/candidate-evidence.json --json
```

Guardar el sidecar fuera del checkout evita introducir un archivo nuevo en el candidato. Se exige limpieza tracked y untracked no ignorados. La única excepción exacta es `.lufy/config/project.yaml` untracked, cuyo contenido queda ligado por hash; una versión tracked modificada se rechaza como dirty. Se rechazan flags `assume-unchanged` y `skip-worktree` porque pueden ocultar cambios. Si el checkout contiene otros internos locales nuevos, usar un checkout limpio aparte. No borrar ni ocultar esos archivos para forzar una aceptación.

El schema JSON `candidate-evidence/v1` conserva hashes del repositorio, políticas y evidencia; OIDs de commits/árboles de base y candidato; digest del diff raw de paths/modos/object IDs; y rol del claim. No contiene rutas privadas, bodies de review ni credenciales. Las políticas ligadas son `.lufy/contracts/delivery.md`, `.lufy/contracts/result-contract.md` y presencia/contenido de `.lufy/config/project.yaml`. Las dos políticas canónicas deben estar tracked en el candidato. Son archivos regulares y no se admiten symlinks en sus rutas. El rol se evalúa con las políticas existentes de Result Contract.

Cambiar HEAD aunque mantenga el mismo árbol, mover la base, editar contenido/modos/paths, cambiar evidencia o políticas invalida la comparación. La identidad de repositorio usa el git-common-dir canónico hasheado: worktrees del mismo repositorio comparten identidad y una reubicación se invalida conservadoramente. Un sidecar mal formado, duplicado, con claves ajenas o mayor de 64 KiB se rechaza.

Las lecturas Git usan argv literal, un plazo de observación de 10 segundos, hasta 1 segundo adicional de cleanup de pipes y salida máxima de 2 MiB por comando. Se deshabilitan optional locks, fsmonitor, hooks, external diff y textconv; los overrides `GIT_*` heredados no pueden desviar el target. Antes de consultar status se rechaza cualquier configuración `filter.*` efectiva, incluyendo includes locales, y cualquier gitlink/submódulo: esos mecanismos pueden ejecutar procesos durante una lectura aparentemente inocua. Los tests usan canaries sintéticos para comprobar que el rechazo ocurre antes de ejecutarlos. Los [filtros de Git](https://git-scm.com/docs/gitattributes) se tratan como configuración ejecutable, no como texto pasivo.

No se crea un árbol sintético ni se escribe el índice. Dos observaciones detectan cambios visibles durante la captura; no constituyen un lock contra un mutador hostil concurrente. El repositorio y su configuración deben permanecer estables durante la operación. Verificar de nuevo inmediatamente antes de consumir la evidencia.

Este binding identifica el snapshot declarado. No autentica autoría ni demuestra que los comandos del contrato se ejecutaron sobre ese snapshot. Los archivos ignorados, dependencias externas y estado fuera de Git no forman parte del árbol comprometido. La identidad no reemplaza pruebas, review, autorización ni `result-transition/v1`, y no transforma la telemetría del bloque 4 en un gate.

La demostración del bloque 5 usó un clon local limpio del HEAD publicado `146b1a6`, un contrato de review válido y un sidecar externo: verify devolvió `current`, una edición temporal de README fue rechazada con exit 4, y restaurar sus bytes devolvió `current`. Los tests agregan invalidación por commit distinto aun con igual árbol, cambio de base/rol/evidencia/repositorio, políticas/config/modos y flags de índice que ocultan cambios. Esta demostración verifica el mecanismo de binding; no atribuye al contrato una validación histórica sobre código que todavía no existía.
