## ADDED Requirements

### Requirement: Routing de implementación por dominio
El `orchestrator` SHALL enrutar tareas de implementación a agentes especializados por dominio cuando la solicitud tenga señales claras de frontend, backend, mobile o arquitectura de microservicios.

#### Scenario: Routing frontend
- **WHEN** una tarea requiere cambios en UI web, componentes, routing frontend, estado cliente, accesibilidad visual o integración de interfaces web
- **THEN** el `orchestrator` enruta la implementación a `frontend-developer` con objetivo, alcance, archivos relevantes y criterios de validación

#### Scenario: Routing backend
- **WHEN** una tarea requiere APIs, servicios server-side, persistencia, auth, jobs, colas o backend Go/Node/Python
- **THEN** el `orchestrator` enruta la implementación a `backend-developer` con contrato esperado, restricciones del repo y validación requerida

#### Scenario: Routing mobile
- **WHEN** una tarea requiere React Native, Flutter, iOS, Android, mobile CI, native modules o comportamiento offline mobile
- **THEN** el `orchestrator` enruta la implementación a `mobile-developer` con plataformas objetivo y restricciones conocidas

#### Scenario: Routing de arquitectura de microservicios
- **WHEN** una tarea requiere definir límites de servicios, comunicación distribuida, eventos, Kubernetes, resiliencia o service mesh
- **THEN** el `orchestrator` consulta `microservices-architect` para diseño/handoff antes de asignar implementación de código cuando corresponda

### Requirement: Fallback seguro a implementer
El `orchestrator` SHALL mantener `implementer` como fallback para tareas que no encajan claramente en un dominio especializado.

#### Scenario: Tooling, OpenSpec, docs o CI
- **WHEN** una tarea toca CI, scripts, OpenSpec, documentación, configuración, `.opencode`, agentes, installer local o cambios pequeños no dominiales
- **THEN** el `orchestrator` usa `implementer` en vez de forzar un agente de dominio

#### Scenario: Dominio ambiguo
- **WHEN** una tarea no tiene dominio claro o toca múltiples dominios sin alcance acotado
- **THEN** el `orchestrator` usa `explorer` primero o pide una decisión breve antes de implementar

### Requirement: Agentes de dominio adaptados al repo
Los agentes especializados SHALL respetar las reglas locales del repositorio y no SHALL asumir mecanismos externos inexistentes.

#### Scenario: Sin context-manager externo
- **WHEN** un agente de dominio necesita contexto del proyecto
- **THEN** obtiene contexto leyendo archivos y artefactos del repo, sin requerir un `context-manager` externo

#### Scenario: Sin delivery desde agentes de dominio
- **WHEN** un agente de dominio completa una implementación
- **THEN** reporta cambios, evidencia y riesgos sin hacer commit, push, PR ni sync de GitHub Projects

#### Scenario: Validación real
- **WHEN** un agente de dominio reporta validación
- **THEN** incluye comandos reales ejecutados y resultados, o declara explícitamente qué validación no estuvo disponible

### Requirement: Gates posteriores conservados
El `orchestrator` SHALL conservar la separación entre implementación, validación, revisión y delivery.

#### Scenario: Validación posterior
- **WHEN** un agente especializado termina cambios con necesidad de evidencia independiente
- **THEN** el `orchestrator` puede llamar a `validator` para compile/test evidence sin editar archivos

#### Scenario: Revisión posterior
- **WHEN** una implementación tiene riesgo de calidad, seguridad, arquitectura o cobertura
- **THEN** el `orchestrator` puede llamar a `reviewer` para findings y merge recommendation

#### Scenario: Delivery autorizado
- **WHEN** el usuario autoriza explícitamente commit, push, PR o sync
- **THEN** el `orchestrator` enruta esas operaciones a `delivery`, no a agentes de dominio
