# Memoria, skills y contexto

Son capacidades locales para reducir redescubrimiento. No son un preflight obligatorio para cada pedido ni reemplazan archivos, comandos o instrucciones explícitas del usuario.

## Memoria Obsidian

La configuración generada usa `.lufy/memory`, con política Git `ignored`. Las notas son Markdown portable; el índice de backlinks es derivado. La inicialización crea un `.gitignore` dentro de la memoria que conserva algunos archivos de orientación e índice; `ignored` no implica que todo `.lufy/` esté excluido del Git del proyecto. Revisá configuración y datos existentes antes de inicializar.

```sh
lufy-ai memory init --target /ruta/al/proyecto
lufy-ai memory status --target /ruta/al/proyecto --json
lufy-ai memory search --target /ruta/al/proyecto 'ownership'
lufy-ai memory capture --target /ruta/al/proyecto --title 'Conservar archivos propios' --type lesson 'La ubicación en un directorio gestionado no concede ownership sobre archivos ajenos.'
lufy-ai memory validate --target /ruta/al/proyecto --json
```

Capturá decisiones, reglas, flujos, conceptos o aprendizajes durables; no resúmenes de cada turno ni secretos. Una nota de tipo `decision` debe explicar el porqué. Buscá primero para evitar duplicados. `memory connect --target <repo> <slug-origen> <slug-destino>` conecta notas existentes; `--bidirectional` permite ambas direcciones. Si editás enlaces manualmente, reconstruí con `memory index` y validá.

Usá hallazgos compactos con path, estado y relevancia. La memoria puede quedar desactualizada: contrastá sus afirmaciones con la fuente actual. No se borran notas personales por cambiar el catálogo del producto.

## Registro de skills

```sh
lufy-ai skills status --target /ruta/al/proyecto --json
lufy-ai skills ensure --target /ruta/al/proyecto
```

`ensure` reconstruye el registry cuando falta o está stale; `refresh` fuerza el escaneo. El registro permite resolver paths exactos de `SKILL.md`. El catálogo instala cinco skills: `lufy-check` y las cuatro `lufy-mem-*`. No instala cadenas de roles.

## Grafo lexical

```sh
lufy-ai context scan --target /ruta/al/proyecto --json
lufy-ai context build --target /ruta/al/proyecto --json
lufy-ai context status --target /ruta/al/proyecto --json
lufy-ai context query --target /ruta/al/proyecto --json 'ownership'
lufy-ai context diff --target /ruta/al/proyecto --base origin/develop --json
```

`scan` inspecciona sin persistir; `build` escribe grafo, resumen y manifest bajo `.lufy/context`. El extractor v3 indexa archivos, símbolos y referencias lexicales. No es un análisis semántico completo ni calcula cobertura scenario-task-test. Los índices anteriores quedan stale y se reconstruyen desde fuentes.

`path` y `explain` ayudan a inspeccionar relaciones. `context review` combina diff Git y límites declarados; no ejecuta pruebas ni aprueba el cambio. Las exclusiones de archivos sensibles reducen la ingesta, pero no equivalen a un escáner universal de secretos. Un estado `stale` requiere reconstrucción sólo si el pedido necesita ese índice; no impide leer las fuentes directamente.
