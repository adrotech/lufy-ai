# Configuración local de Lufy

- `config/project.yaml`: perfil del proyecto y ejecución inline (schema 2).
- `contracts/work.md`: autonomía, delegación y comprobación del resultado.
- `contracts/delivery.md`: autorización y evidencia para entrega.
- `managed-state/install-state.json`: manifest de assets gestionados (schema 3).
- `skill-registry.json`: capacidades instaladas; regenerable con `lufy-ai skills ensure`.
- `memory/`: memoria portable Obsidian cuando está habilitada.

Ejemplo mínimo:

```yaml
schema_version: 2
tool: codex
execution:
  mode: inline
  delegation: auto
```

`auto` permite delegar si aporta valor; `on_request` sólo ante un pedido; `off` mantiene el trabajo en el principal. Las instrucciones explícitas del usuario y los permisos del host prevalecen. Las configuraciones y manifests anteriores se rechazan; no se convierten ni interpretan mediante compatibilidad.
