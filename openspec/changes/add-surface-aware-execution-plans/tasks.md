## 1. Contrato y resolución

- [x] Añadir modelos neutrales, puertos y errores accionables.
- [x] Implementar `SurfaceResolverStrategy` con precedencia, roots específicas y composición.
- [x] Cubrir selección explícita, automática, empate, archivos compartidos y ambigüedad.

## 2. Validación y capacidades

- [x] Añadir `capabilities` opcionales a `ProjectSurface` preservando rescan/TUI.
- [x] Implementar `ValidationPlanFactory` para expectativas, comandos, señales y capacidades.
- [x] Cubrir frontend, backend, contrato fullstack, persistencia y realtime/offline.

## 3. Adapters y caso de uso

- [x] Implementar lectura de project config y cambio Git mediante puertos.
- [x] Orquestar plan read-only con dependencias inyectables.
- [x] Validar paths y normalizar entradas de CLI.

## 4. CLI y TUI

- [x] Exponer `lufy-ai plan` humano/JSON con flags de superficie, base, archivos y capacidades.
- [x] Integrar ayuda, command palette y tests de contrato CLI.

## 5. Documentación y validación

- [x] Documentar arquitectura, ejemplos y consumo por agentes.
- [x] Ejecutar validación Go agrupada y OpenSpec strict.
- [x] Revisar diff, compatibilidad YAML y estado final de tareas.
