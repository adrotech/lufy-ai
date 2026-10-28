# Lufy: trabajo inline centrado en resultados

El agente principal comprende el pedido, consulta el contexto pertinente, ejecuta y comprueba el resultado. La política canónica es [.lufy/contracts/work.md](.lufy/contracts/work.md).

La configuración vigente vive en `.lufy/config/project.yaml`, con `schema_version: 2`, `tool: codex` y `execution.mode: inline`. `execution.delegation` admite `auto`, `on_request` y `off`. Esa elección controla delegación; los permisos efectivos del host y la autorización del usuario siguen siendo obligatorios.

Usa el perfil de superficies, stacks y comandos de validación del proyecto cuando corresponda. Para investigación contrasta fuentes; para escritura verifica fidelidad y exactitud; para código prueba comportamiento y regresiones; para datos conserva consultas reproducibles. No generes artefactos o pruebas que no contribuyan al resultado.

Las capacidades locales se descubren en `.lufy/skill-registry.json` y `.agents/skills`. La memoria aporta pistas históricas y no sustituye evidencia actual. El único hook instalado orienta la sesión con metadata local.

Para operaciones de entrega aplica [.lufy/contracts/delivery.md](.lufy/contracts/delivery.md). No hay una selección de metodología por tier ni un circuito obligatorio de roles.
