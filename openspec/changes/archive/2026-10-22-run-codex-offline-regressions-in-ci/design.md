# Diseño

El Quality gate existente mantiene el check agregado obligatorio. El comando local/CI ejecuta unittest estándar; apt declara python3 en Ubuntu. macOS/Windows conservan sus pruebas Go. No se ejecutan modelos ni se solicita autenticación.

No añadir tests que sólo reflejen YAML: las regresiones existentes son la prueba funcional. Revisar filtros y propagación estáticamente y mediante los checks del PR. Validación agrupada con scripts/validate.sh, OpenSpec estricto, paridad y guard/whitespace del rango PR. Fuente y mirror se mantienen visibles. Sin cambios OpenCode/Pi.
