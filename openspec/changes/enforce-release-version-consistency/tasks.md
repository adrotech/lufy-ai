## 1. Fuente canonica local

- [x] 1.1 Extender el checker con version esperada, changelog, binario y root inyectable.
- [x] 1.2 Agregar setter coordinado de version para documentos copiables.
- [x] 1.3 Agregar pruebas shell de casos positivos y negativos.
- [x] 1.4 Actualizar source tree y changelog a la version planificada `v0.6.22`.

## 2. Gates remotos

- [x] 2.1 Validar source/changelog contra el tag calculado antes de crear tags automaticos.
- [x] 2.2 Validar source/changelog sobre el checkout del tag en `release.yml`.
- [x] 2.3 Verificar version exacta del artifact construido y publicado.
- [x] 2.4 Configurar `cache-dependency-path` para el modulo Go anidado en todos los workflows afectados.

## 3. Validacion agrupada

- [x] 3.1 Ejecutar pruebas enfocadas del versionado.
- [x] 3.2 Ejecutar `scripts/validate.sh` y revisar el diff contra `origin/develop`; el runner local se detiene porque no tiene Go y la suite completa queda a cargo del CI del PR.
- [x] 3.3 Registrar que `openspec validate` no esta disponible en el runner local; la estructura se reviso manualmente contra el workflow del repositorio.
