# Mejora del evaluador, 29/10/2026

Estado: implementada y validada localmente en `codex/improve-inline-evaluator`, desde develop `64303ff3f2c03e895b5f8b83c66faf913113320d`. Commit, push y PR hacia develop autorizados el 29/10/2026. Ejecución inline, sin modelos ni código de candidatos.

## Cambio y aceptación

El clasificador anterior confundía operadores escritos entre comillas con efectos del shell y desconocía lecturas frecuentes. `command_effects.py` agrega una gramática POSIX acotada que conserva esa distinción, reconoce checksums y `git status` con opciones limitadas. Un diff Git sólo se acepta como lectura con pager, helpers externos y textconv desactivados. Configuración explícita, opciones y construcciones fuera de la gramática permanecen desconocidas; las mutaciones y delivery identificables mantienen su clasificación.

`assess` y el auditor incluyen versión y diagnósticos por evento, sin duplicar argumentos de comandos. El helper participa en los fingerprints del runner y de la comparación: cambiarlo invalida la estabilidad de fuentes observadas. No hay nueva dependencia externa, proceso del producto ni requisito para el usuario.

Criterios comprobados: argumentos con operadores literales se leen correctamente; escrituras/delivery en secuencias no quedan ocultos; scripts sed, opciones Git peligrosas, expansiones, pipes, heredocs y Python arbitrario no obtienen aprobación por descarte. Hay límites explícitos de tamaño, tokens y profundidad. Un texto sigue en `needs_review`; el código necesita su oráculo. El auditor conserva el veredicto original y las métricas humanas ausentes.

## Cobertura observada

Lectura offline de 24 registros originales, con integridad de los hashes de trazas y sin modificar registros:

| Clasificación | Antes | Ahora |
| --- | ---: | ---: |
| Lectura reconocida | 42 | 55 |
| Comprobación declarada | 4 | 4 |
| Comando desconocido | 33 | 20 |

Se reconocen 13 comandos adicionales. Los 20 restantes incluyen 12 heredocs, dos expansiones, cuatro comandos no soportados y dos diffs sin desactivación explícita de helpers. [Resumen y hashes de fuentes](mejora-evaluador-2026-10-29-coverage.json).

Esto mide cobertura del clasificador. No reevalúa la calidad de las entregas, no reejecuta oráculos ni modifica los veredictos de la comparación. La conclusión histórica sobre ventaja general y costo humano sigue inconclusa.

## Validación ejecutada

- RED: los casos nuevos expusieron el desconocimiento de lecturas comunes y la confusión de argumentos entre comillas en la implementación anterior.
- `python3 -m unittest discover -s tools/codex-behavior-pilot -p test_command_effects.py`: nueve pruebas focalizadas aprobadas, con casos positivos y adversariales.
- `GOCACHE=/private/tmp/lufy-inline-go-cache scripts/validate.sh`: aprobado; **63 pruebas Python**, tests Go con **80,6% de cobertura**, vet, build, whitespace contra origin/develop, PR guard, versiones, YAML y paridad de assets. [Log](mejora-evaluador-2026-10-29-validation.log).
- `python3 /private/tmp/lufy-evaluator-20261029/replay.py`: 79 comandos de 24 registros, hashes y archivos de evidencia conservados. La función baseline se obtiene del commit citado; no se ejecutan las instrucciones de las trazas.
- Revisión estática de archivos modificados y nuevos: separación entre sintaxis y efectos, procedencia del helper, preservación de revisión editorial/oráculo y contenido histórico.

ShellCheck no estaba instalado y se informó omitido. Este informe registra la validación local previa al PR; los resultados remotos se informan en GitHub. Las pruebas no acreditan binarios, entorno, configuración ni permisos efectivos: esta gramática no es un sandbox. El checkout original, su código local y el ejecutable instalado se preservaron.
