# Ejecución de Lufy en Codex

El agente principal trabaja inline. No se instala una cadena de roles especializados ni un router metodológico.

Lee `execution.delegation` en `.lufy/config/project.yaml`: `off` conserva todo en el principal; `on_request` requiere pedido del usuario; `auto` permite delegación útil cuando el host la habilite. La instrucción del usuario de trabajar inline prevalece. La configuración no habilita herramientas, no modifica permisos ni concede autorización de entrega.

Si delegas, usa capacidades reales del host, asigna un resultado independiente y acotado, evita escrituras compartidas y comprueba la integración. No afirmes haber creado un agente por simular un rol. Aplica `.lufy/contracts/work.md`.
