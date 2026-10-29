# Piloto del trabajo inline

Piloto de la [propuesta de evolución](../../docs/propuestas/lufy-inline-2026-10-24/README.md), iniciado en el bloque 1. Los bloques de implementación ya están integrados en `develop`; la instalación personal se verificó por separado. Este piloto conserva sus fixtures aislados; el [bloque 5](FAMILIES.md) audita evidencia por familias sin reescribir las ejecuciones anteriores.

La fuente neutral es `.lufy/contracts/work.md`. Su copia embedded distribuye el mismo contenido; no es otra fuente editorial. Cada fixture recibe sólo esa política y los límites de su caso en `AGENTS.md`, sin roles, contratos de delivery antiguos ni instrucciones metodológicas contradictorias.

## Casos y aceptación

`work-scenarios.json` contiene seis pedidos sintéticos: explicar código, proponer una funcionalidad, redactar una guía, traducirla, analizar y corregir un filtro, y diagnosticar duplicación por un join.

`work.py` observa finalización del turno, respuesta textual vinculada a la traza, comandos soportados, intentos de escritura y diferencias del filesystem. Sólo permite cambios en los archivos declarados por cada fixture. Los intentos de mutación read-only y de delivery se rechazan aunque fallen; los eventos/comandos desconocidos quedan inconclusos. El clasificador admite una gramática acotada de comandos y secuencias simples: no es un parser general de shell ni un sandbox.

El caso `fix-filter` tiene un oráculo funcional externo al fixture que prueba entradas distintas de las pruebas visibles. El código inicial debe fallar; la corrección debe filtrar por cualquier estado, conservar orden y datos, admitir `None` y entradas vacías. La comprobación no acepta un exit cero previo a las assertions, symlinks ni cambios de archivos durante la ejecución del oráculo. No se acredita éxito por la afirmación del agente.

El candidato se ejecuta mediante `codex sandbox` con `sandbox_mode="read-only"`, sin pasar autenticación ambiental al proceso de pruebas. Si ese sandbox no puede iniciar, el resultado es inconcluso: no hay fallback al proceso del host. El caso de implementación verifica disponibilidad del oráculo antes de invocar el modelo. Las pruebas offline ejecutan únicamente sus propias fuentes sintéticas confiables, no código producido por modelos.

En análisis, propuesta y redacción, pasar los controles de alcance y ejemplos literales produce `needs_review`, no una aprobación de calidad automática. Se incluyen criterios para contrastar el contenido. Ese estado pertenece al experimento y no obliga a pedirle aprobación al usuario en el producto: el evaluador o revisor del ensayo documenta qué pudo comprobar. No se aplica cobertura de código a un texto.

## Clasificación de comandos

`command_effects.py` implementa `command-effects/v2`: distingue operadores del shell de argumentos literales entre comillas. Reconoce secuencias simples, lecturas, checksums, `git status` con opciones acotadas y `git --no-pager diff --no-ext-diff --no-textconv` sin helpers. La configuración Git explícita, opciones desconocidas, pipes, expansiones, heredocs y Python arbitrario continúan inconclusos; los comandos de comprobación declarados se reconocen por argv exacto. Las mutaciones y delivery conocidos conservan su clasificación aunque aparezcan en una secuencia.

Cada evaluación registra `classifier_version` y `command_observations` con índice del evento, clase y motivo; no duplica los argumentos del comando en ese diagnóstico. El helper participa en los fingerprints de procedencia del runner y de la comparación. Una lectura reconocida sólo despeja esa incertidumbre de alcance: un texto sigue requiriendo revisión semántica y un cambio de código sigue necesitando su oráculo.

La clasificación es estática y acotada. No ejecuta comandos, no atesta binarios, configuración, entorno ni permisos efectivos y no sustituye el aislamiento. Las construcciones fuera de su gramática no se aceptan por descarte. Los resultados históricos se conservan; cualquier reevaluación identifica la versión actual por separado.

## Uso

Listar casos no ejecuta modelos, carga autenticación ni crea evidencia:

```bash
python3 tools/codex-behavior-pilot/work.py
```

Pruebas offline, incluidas por el discovery existente de `scripts/validate.sh`:

```bash
python3 -m unittest discover -s tools/codex-behavior-pilot -p 'test_work.py' -v
```

La corrida real es opt-in, requiere un ejecutable compatible y un directorio externo nuevo. Usa la autenticación existente de Codex, sin copiar credenciales ni cambiar configuración global:

```bash
python3 tools/codex-behavior-pilot/work.py --live --case fix-filter \
  --codex /ruta/al/codex --output /ruta/externa/nueva --timeout 60
```

Las opciones observadas y las restricciones de ejecución deben contrastarse con `codex exec --help` de ese ejecutable. `--model` es opcional; no se inventa identidad del modelo servido. El piloto solicita `workspace-write`, sin aprobaciones interactivas, y desactiva agentes, apps, hooks y plugins sólo en esa invocación. Esas solicitudes no atestan por sí solas el aislamiento efectivo. No se usan flags de bypass ni se eliminan restricciones heredadas del host.

No se requiere JSON de salida al agente. El runner guarda localmente prompt, respuesta, JSONL, stderr, snapshots, hashes del binario y fuentes, oráculo y resumen. Las transiciones del producto no dependen del JSON del experimento. El primer fallo de infraestructura o evidencia inconclusa detiene nuevas invocaciones; los casos restantes quedan `not_run`. Un cambio de fuentes/binario invalida la aceptación del caso.

Las trazas permanecen locales. No deben publicarse automáticamente: aunque el fixture sea sintético, el runtime puede añadir contexto propio. Los ensayos reales y sus límites se registran en el informe del bloque; no se mezclan versiones del evaluador ni se repiten muestras sólo para conseguir verde.

Interfaces contrastadas con la [documentación oficial de ejecución no interactiva](https://learn.chatgpt.com/docs/non-interactive-mode) y [referencia de CLI](https://learn.chatgpt.com/docs/cli/reference), además del help del ejecutable observado.

## Límite del bloque

Esto implementa fixtures, evaluador y runner. No acredita paralelismo nativo, una comparación causal de metodologías ni ahorro de recursos. La configuración, catálogo e instalación actuales ya usan inline; la eliminación general de módulos retirados se completó en el bloque 6. El baseline de comparación se conserva en Git, sin ser un requisito de compatibilidad del producto.
