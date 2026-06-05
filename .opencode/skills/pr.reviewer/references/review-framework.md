# Framework De Review Agnóstico

Checklist base para revisar PRs sin depender de lenguaje o framework.

## Arquitectura Y Diseño

- Separación clara entre entrada/adaptadores, aplicación/orquestación, dominio/reglas e infraestructura/dependencias.
- Dependencias apuntan hacia abstracciones estables; infraestructura no fuerza reglas de negocio.
- Handlers/controllers/resolvers/consumers son delgados y delegan reglas.
- Contratos públicos no cambian sin migración, compatibilidad o comunicación explícita.
- Scopes transaccionales son estrechos y consistentes con el riesgo del flujo.
- No se agregan acoplamientos innecesarios, singletons globales o dependencias difíciles de testear.
- Nuevas abstracciones tienen un caso de uso real; no hay sobreingeniería.

## Correctitud Funcional Y Contratos

- Inputs se validan antes de ejecutar efectos secundarios.
- Estados inválidos y transiciones imposibles están bloqueados explícitamente.
- Errores conservan contexto sin filtrar datos sensibles.
- Flujos con retry o eventos son idempotentes.
- Cambios backward-incompatible están documentados y versionados.
- Antes/después del comportamiento está claro para escenarios relevantes.
- Datos opcionales, nulos, vacíos y límites se manejan de forma explícita.

## Seguridad Y Privacidad

- No hay secretos hardcodeados ni tokens en logs, config o tests.
- Autenticación/autorización no se debilita.
- Inputs externos no llegan sin sanitización a consultas, comandos, templates o serialización peligrosa.
- PII y datos sensibles no se registran ni se exponen en errores.
- Dependencias externas, callbacks y webhooks validan origen/contrato cuando aplica.
- Permisos nuevos son mínimos y justificados.

## Pruebas Y Evidencia

- Tests cubren comportamiento nuevo, rutas de error y bordes relevantes.
- Tests no dependen de orden global, tiempo real innecesario o estado compartido frágil.
- Mocks/fakes verifican contratos importantes, no solo llamadas triviales.
- Hay evidencia real de validación o se declara claramente que falta.
- Cambios riesgosos tienen pruebas de integración, contrato o simulación suficiente.
- Fixtures y snapshots son legibles y no ocultan comportamiento esencial.

## Observabilidad Y Operación

- Logs estructurados incluyen contexto útil sin PII.
- Métricas/trazas cubren rutas críticas, errores y latencia cuando aplica.
- Alertas/monitores se actualizan si cambia un flujo operativo relevante.
- Errores de dependencias son distinguibles para diagnóstico.
- No hay log spam en loops ni cardinalidad explosiva en métricas.
- Timeouts, retries y backoff son configurables cuando el riesgo lo requiere.

## Migraciones, Configuración E Infraestructura

- Migraciones son reversibles o tienen plan de rollback.
- Cambios de schema preservan compatibilidad con versiones en despliegue.
- Nuevas variables/configs tienen defaults seguros y documentación.
- Infra/CI/CD no rompe ambientes existentes.
- Se evita cambiar puertos, auth defaults, rutas públicas o contratos sin autorización.

## Mantenibilidad Y Complejidad

- Funciones/métodos tienen responsabilidad clara y tamaño razonable para el contexto.
- Nombres reflejan intención de negocio/técnica.
- Complejidad condicional se reduce con guard clauses o composición cuando ayuda.
- Código muerto, comentarios obsoletos y TODOs sin seguimiento no se introducen.
- Se evita duplicación que pueda divergir en reglas críticas.

## Concurrencia, Consistencia Y Resiliencia

- Operaciones concurrentes protegen estado compartido.
- Efectos secundarios tienen orden seguro ante fallas parciales.
- Transacciones y locks cubren exactamente el alcance necesario.
- Reprocesamiento no duplica efectos.
- Fallas de dependencias dejan el sistema en estado observable y recuperable.

## Señales De Riesgo Para Elevar Severidad

- Cambio toca dinero, permisos, datos personales, migraciones, auth, contratos públicos o procesamiento masivo.
- PR grande mezcla muchos objetivos sin slices claros.
- La evidencia de pruebas no cubre el comportamiento modificado.
- Hay comentarios previos sin resolver sobre el mismo riesgo.
- El diff agrega lógica crítica en capas de entrada o infraestructura sin tests.
