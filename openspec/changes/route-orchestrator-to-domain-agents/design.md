## Context

El `orchestrator` actual es `mode: primary`, no edita archivos ni ejecuta shell, y enruta a agentes especializados con permisos `task`. El `implementer` actual es un subagente de edición acotada, sin permiso para delegar (`task: "*": deny`) y con la regla explícita de no delegar a otros agentes.

Los subagents externos propuestos son útiles como punto de partida, pero no deben copiarse literalmente: asumen `context-manager`, stacks y toolchains que no necesariamente existen en este repo. Deben adaptarse al modelo local: `AGENTS.md`, OpenSpec, delivery separado, validación real y documentación humana en español.

## Goals / Non-Goals

**Goals:**

- Hacer que `orchestrator` pueda enrutar tareas de implementación a especialistas de dominio.
- Mantener `implementer` como fallback seguro para cambios no cubiertos por dominio.
- Definir señales de routing claras para frontend, backend, mobile y microservices architecture.
- Establecer un formato de handoff desde `orchestrator` hacia agentes especializados.
- Mantener `validator`, `reviewer` y `delivery` como gates separados.
- Adaptar los agentes externos al repo sin asumir `context-manager` ni toolchains inexistentes.

**Non-Goals:**

- No remover `implementer` en esta propuesta.
- No convertir `implementer` en dispatcher interno.
- No permitir delivery desde agentes de dominio.
- No añadir dependencias externas ni automatización remota.
- No implementar cambios de producto; solo workflow/agentes.

## Decisions

1. **Routing vive en `orchestrator`, no en `implementer`**

   `orchestrator` ya tiene la responsabilidad de coordinar roles. Poner routing dentro de `implementer` rompería su boundary actual y haría más difícil controlar permisos.

2. **`implementer` se mantiene como fallback**

   El repo necesita un ejecutor para CI, scripts, OpenSpec, docs, configuración, `.opencode`, installer Go y cambios pequeños. Los especialistas de dominio no cubren bien esos casos.

3. **Agentes externos se adaptan, no se copian literalmente**

   Cada agente nuevo debe eliminar supuestos como `context-manager`, ajustar permisos, reducir ambición, respetar validación agrupada y entregar reportes en el formato local.

4. **Reglas de routing explícitas**

   - `frontend-developer`: UI web, componentes, routing frontend, client state, accesibilidad visual.
   - `backend-developer`: APIs, servicios server-side, persistencia, auth, jobs, backend Go/Node/Python.
   - `mobile-developer`: React Native, Flutter, iOS, Android, mobile CI o native modules.
   - `microservices-architect`: límites de servicios, comunicación distribuida, eventos, Kubernetes, resiliencia, service mesh; normalmente diseña primero y no reemplaza un ejecutor de código.
   - `implementer`: tooling, CI, scripts, OpenSpec, docs, configuración, agentes, installer local y cambios pequeños no dominiales.

5. **Gates se mantienen independientes**

   Después de implementación, `orchestrator` debe seguir pudiendo llamar `validator` para evidencia y `reviewer` para riesgo/calidad. `delivery` solo se usa con autorización explícita.

## Risks / Trade-offs

- **Routing demasiado agresivo** → Mantener fallback a `implementer` y permitir que especialistas devuelvan `blocked` si el dominio no aplica.
- **Especialistas con supuestos externos** → Adaptar instrucciones y remover dependencias conceptuales como `context-manager`.
- **Duplicación de responsabilidades** → Documentar claramente que `orchestrator` decide, agente especializado implementa, `validator` valida y `delivery` entrega.
- **Pérdida de simplicidad** → Introducir esta capacidad sin borrar `implementer`; evaluar uso después.
- **Cambios grandes sin exploración previa** → Para tareas ambiguas o arquitectónicas, `orchestrator` debe llamar primero a `explorer` o `microservices-architect` en modo diseño.
