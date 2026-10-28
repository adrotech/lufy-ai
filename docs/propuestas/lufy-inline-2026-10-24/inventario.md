# Inventario: conservar, modificar y retirar

Parte de la [propuesta Lufy inline](README.md). Snapshot de referencia: `1cc18d4647afc61ea1f6b1cb726ba74128fded5f` (`origin/develop`, `v0.6.25`). Los paths identifican las áreas de responsabilidad, no una lista exhaustiva de archivos que se editarán. Este documento no elimina ni modifica los componentes enumerados.

## 1. Regla de disposición

Cada componente tiene un destino y un motivo. «Retirar» significa que deja de formar parte del producto nuevo en el alcance indicado. La implementación deberá distinguir assets gestionados, archivos modificados por el usuario y contenido que sólo pertenece al proyecto.

Lufy sólo tiene un usuario y no necesita retrocompatibilidad. Un componente no se conserva para atender instalaciones anteriores, consumidores hipotéticos o adapters históricos. Las referencias internas se actualizan o eliminan junto con él. Permanece únicamente el código que aporta una capacidad elegida para el producto actual.

## 2. Comportamiento e instrucciones

| Elemento actual | Destino | Modificación concreta | Condición de aceptación |
| --- | --- | --- | --- |
| `.lufy/contracts/work.md` y mirror embedded | Fuente nueva de comportamiento | Política neutral del bloque 1; seleccionarla desde el catálogo/agente principal en el bloque 2. El piloto ya la carga como única política del fixture. | Paridad de distribución y selección sin instrucciones metodológicas contradictorias. |
| Clasificación obligatoria T1/T2/T3 | Retirar del camino inline | Eliminar la selección de workflow por tier de las instrucciones nuevas y del catálogo Codex migrado. | Una tarea arquitectónica puede recibir una propuesta rigurosa sin que se exija OpenSpec; una edición mecánica grande no activa un workflow por cantidad de archivos. |
| `.codex/agents/sdd-router.toml` | Eliminar | Borrar definición, asset embebido, binding y referencias obligatorias. No dejar alias ni lector de ese rol. | Inline no invoca router ni fabrica un handoff para continuar. |
| `.codex/agents/orchestrator.toml` | Retirar como coordinador obligatorio | El agente principal del host ejecuta; el agente read-only existente no gana permisos. | Se pueden completar unidades acotadas sin saltos de rol, respetando sandbox y restricciones del host. |
| Roles de explorer, implementer, test-writer, validator y reviewer | Convertir en capacidades opcionales | Mantener especialización que aporte valor; no instalar una cadena obligatoria. El mismo ejecutor puede investigar, implementar y comprobar. | Un usuario puede pedir explícitamente un revisor independiente; un bug sencillo no exige delegación. |
| Rol delivery | Convertir en capacidad opcional del modelo nuevo | Separar procedimientos Git/GH de la obligación de un proceso/agente adicional; preservar autorización y checks. | Ninguna recomendación, score o siguiente paso constituye permiso para publicar. |
| `.codex/instructions/proportional-workflow.md` | Sustituir en instalación migrada | Reescribir alrededor de resultado, alcance, evidencia y continuidad inline. | No hay instrucciones activas que vuelvan a imponer metodología por tier. |
| `AGENTS.md.template`, bloque gestionado de `AGENTS.md` y `lufy-ia.harness.md` | Reducir y redistribuir | Mantener hechos y restricciones esenciales; retirar repeticiones de workflow y cargar capacidades pertinentes bajo demanda. | Las reglas efectivas son coherentes; el texto humano del proyecto fuera del bloque gestionado queda intacto. |
| `.codex/lufy-agent-mapping.md` | Simplificar | Documentar capacidades realmente disponibles y política de delegación; dejar de presentar emulación de roles como ejecución nativa. | Runtime sin subagentes trabaja inline; con política automática puede repartir ramas independientes sin pedir coordinación al usuario. |
| Requisito de TDD delegado en todos los cambios sustantivos | Retirar como obligación | Preservar pruebas discriminantes y permitir TDD por el ejecutor. Pedir independencia cuando el riesgo/oráculo lo requiera. | Tests no se debilitan para pasar; ausencia de delegación no bloquea por sí sola. |
| Score de review ≥80% y mínimo universal de ocho desk checks | Retirar como gate | Usar aceptación, hallazgos, severidad, riesgos cubiertos y evidencia faltante. | Un defecto crítico bloquea aunque el resto del trabajo sea bueno; la cantidad de casos responde al cambio. |
| Handoffs a uno mismo y YAML completo en respuestas humanas | Retirar | Mantener un resumen útil y referencias sólo cuando haya una transferencia real o consumidor automatizado. | Una explicación o propuesta termina sin metadatos metodológicos vacíos. |
| Aprobación adicional causada solamente por emitir un plan | Retirar del modelo nuevo | Reutilizar autorización del pedido para pasos incluidos en su alcance; preservar decisiones de producto, seguridad y efectos no autorizados. | «Analizá y corregí» puede continuar; «analizá» no habilita corrección. |
| Umbrales universales por número de archivos | Replantear | Usarlos como señales de carga cuando aporten; respetar restricciones explícitas configuradas hasta migrarlas con decisión trazable. | Diez reemplazos mecánicos y una línea de auth reciben evaluación por consecuencias, sin invertir su riesgo por conteo. |

## 3. Configuración, dominio y catálogo

| Ruta o área actual | Destino | Trabajo requerido y frontera |
| --- | --- | --- |
| `tools/lufy-cli-go/internal/core/domain/harness.go` y `routing.go` | Eliminar modelo de tier/metodología | Borrar `MethodologyByTier`, defaults, reglas y traducciones. Conservar `ToolID` o capacidades sólo donde sirvan al diseño actual; actualizar todas sus referencias propias. |
| `internal/projectconfig/` | Reemplazar modelo y parser | Soportar únicamente la configuración actual sin metodología. Trasladar una vez los valores útiles del proyecto; borrar decoders, defaults y tests de formatos anteriores. |
| `internal/harnessconfig/` y `internal/harnesscatalog/` | Simplificar | Resolver capacidades por host sin metodologías. Borrar el catálogo anterior y sus ramas alternativas; no crear una capa de compatibilidad. |
| `internal/cli/harness_flags.go` | Eliminar flags metodológicos | Borrar `--methodology-tier`, parsing, ayuda y tests asociados. No deprecarlo ni conservarlo como alias o no-op. |
| `internal/governance/` | Modificar | Separar restricciones operativas, información del proyecto y selección metodológica. La política nueva no puede seguir exigiendo `MethodologyByTier` indirectamente. |
| `internal/state/state.go` | Reemplazar estado y lector | Regenerar el manifest de la instalación propia con el formato actual; borrar soporte de schemas anteriores. La sustitución usa backup externo y comprobación, no convivencia de estados. |
| `internal/instructions/registry/`, `render/`, `roles/` y bindings | Modificar y depurar | Fuente declarada para instrucciones compartidas; proyecciones consistentes por host; retiro de bindings que sólo conectaban tier, rol y metodología. No sumar una segunda fuente editorial. |
| `internal/codexsurface/`, `codexruntime/` y `codexlifecycle/` | Conservar y adaptar donde corresponda | Verificar capacidades, límites y configuración reales de Codex; distinguir recomendación textual de enforcement del host. |
| `internal/platform/`, `ports/` y adapters de herramientas | Conservar o simplificar según utilidad actual | Reutilizar fronteras útiles y borrar las que sólo sirvan al modelo retirado. No mantener una interfaz por consumidores históricos; actualizar o eliminar sus llamadas propias. |

Los tipos nuevos deben aparecer donde sean necesarios para representar configuración o integración reales. No se crea por anticipado un paquete `work engine`, una DSL de procesos ni un endpoint de planificación que cada pedido deba consultar.

## 4. Motores, skills y datos metodológicos

| Elemento | Destino final | Eliminación y tratamiento de datos |
| --- | --- | --- |
| Adapter de metodología OpenSpec y `internal/opsx/` | Eliminar | Borrar implementación, assets, dependencias, comandos y tests exclusivos. Actualizar o eliminar referencias de cualquier adapter propio afectado; no conservar el motor por ellas. |
| Adapter `lufy-sdd` e `internal/lufysdd/` | Eliminar | Borrar implementación, assets, acciones, templates y tests exclusivos; retirar imports y campos que los seleccionaban. |
| Metodología `none` como excepción para trabajar directamente | Retirar del modelo nuevo | Inline es el camino predeterminado; no una excepción codificada mediante una metodología llamada `none`. |
| Skills `sdd-workflow`, `openspec-*` y `lufy-sdd-*` | Eliminar del producto | Borrar catálogo, fuentes y mirrors gestionados. Su posible uso externo futuro no justifica incluirlas ni reinstalarlas. |
| Slash commands `opsx-*` y acciones SDD de Lufy | Retirar del nuevo conjunto activo | Migración informa qué comandos desaparecen y cómo continuar los documentos existentes. No redirige sus nombres silenciosamente a otro proceso. |
| `openspec/changes`, `openspec/specs` y archivo histórico del proyecto | Extraer conocimiento útil y eliminar estructura obsoleta | Trasladar decisiones e invariantes vigentes a documentación actual. Borrar specs, cambios y checklists sin valor actual; Git/backup conservan el pasado. El trabajo pendiente se identifica como vigente, sustituido o descartado, nunca como implementado por eliminación. |
| `.lufy/workflows/sdd/changes`, decisiones y verificaciones existentes | Depurar y retirar del modelo activo | Reubicar sólo conocimiento y trabajo que sigan siendo necesarios; borrar el contenedor metodológico y contenido obsoleto. Preservar evidencia útil sin mantener el workflow. |
| Plantillas que duplican propuesta, diseño, tareas, verificación y resultado | Retirar las redundantes del catálogo nuevo | Mantener ejemplos de acuerdos breves o análisis de riesgos sólo si su uso aporta valor; sin exigir un set completo. |
| Specs SDD usadas como tests normativos del producto | Actualizar su alcance | No dejar que fixtures antiguos obliguen al nuevo camino a reproducir la arquitectura retirada. Conservar regresiones de permisos, ownership y datos. |

**Cierre exigido del retiro:** el repositorio, assets embebidos, catálogo, instalación y configuración ya no contienen implementaciones activas de los motores eliminados. No quedan parsers, aliases, bridges ni imports que los reactiven. Los tests del comportamiento retirado se sustituyen por aceptación del nuevo; el baseline existe únicamente como revisión histórica para comparar.

## 5. Infraestructura que se conserva

| Componente | Por qué permanece | Ajuste permitido |
| --- | --- | --- |
| CLI Go en `tools/lufy-cli-go` y `scripts/install.sh` | Distribución consistente y toolchain existente | El script sigue siendo wrapper estricto; sin fallback legacy ni nueva CLI Node. |
| `installer`, `managedcontent`, `managedio`, `assets`, `state`, `syncer`, `verify` | Ownership, hashes, idempotencia y detección de drift | Adaptar catálogo y migración; nunca borrar por patrón todo lo que contenga «SDD». |
| `backup`, `conflictplan`, `merger`, `mergeblock`, `uninstaller`, `upgrade` | Preservación de cambios del usuario y recuperación | Probar retiro de assets, conflictos, rollback y reintento de la migración nueva. |
| `surfaceplan` y capacidad actual de `lufy-ai plan` | Contexto de superficies, contratos y validaciones del proyecto | Conservar por utilidad. Se puede cambiar comando o schema actualizando sus llamadas propias; no conservar firmas antiguas por compatibilidad. Ausencia de superficies no bloquea investigación o escritura. |
| `skillregistry` | Resolver capacidades disponibles y procedencia | Carga bajo demanda y eliminación de defaults metodológicos; sin inventar una skill por cada verbo. |
| Branch safety, checks remotos y release desde `main` | Protegen integración y distribución | Mantener controles reales; separar su necesidad de la etiqueta o rol que los ejecutaba. |

## 6. Componentes que deben reducir su alcance o demostrar valor

| Componente | Propuesta | Criterio de permanencia o retiro |
| --- | --- | --- |
| `resultcontract` y contrato humano asociado | Eliminar v1 y simplificar o retirar implementación | Borrar campos metodológicos, parser y validaciones antiguas. Si queda una integración útil, actualizarla al formato mínimo actual en el mismo cambio; sin bridge v1/v2. Si no queda consumidor elegido, eliminar el componente completo. |
| `resultcontractledger` y `runledger` | Conservar idempotencia y evidencia necesarias; revisar solapamiento | Una sola autoridad de transición por ejecución. Retirar duplicación que no tenga una función distinta comprobada. |
| `candidateevidence` | Conservar para comprobaciones de candidatos y entrega donde corresponda | Una verificación acredita su snapshot. No obliga a crear commits durante cada iteración ni sirve como único modelo de evidencia para textos/datos. |
| `reviewevents` y métricas operativas | Mantener eventos que alimenten decisiones concretas | No inferir «cero defectos» porque falta un evento; retirar métricas sin consumidor y minimizar contenido sensible. |
| `adaptive` y `.lufy/contracts/adaptive-routing.md` | Sacar del camino crítico y reevaluar | No concede permisos ni decide automáticamente el workflow nuevo. Si no mejora la asignación en casos objetivo, retirar del conjunto activo Codex. |
| `contextgraph` | Capacidad opcional, con fallback a archivos/búsqueda | Un grafo stale no bloquea ni se trata como evidencia actual. Su uso debe reducir búsquedas o errores en casos relevantes. |
| `memory` con Obsidian | Capacidad opcional y portable | Archivos actuales e instrucciones del usuario prevalecen. No sumar otra memoria por adoptar una idea de gentle-ai. Captura durable selectiva. |
| Informes de tiempo/costo y observabilidad | Conservar sólo cuando sean útiles para el usuario | Comparar costo de mantenimiento y decisiones habilitadas. El plan no incluye mejoras del Observatory de OpenCode. |

## 7. Orden de eliminación y criterio de seguridad

1. Dejar de exigir las reglas retiradas en el piloto inline aislado.
2. Reemplazar configuración y catálogo, eliminando la ruta metodológica antigua y sus referencias.
3. Preparar sustitución puntual de la instalación propia con inventario, backup, preview y recuperación; sin migrador permanente ni lectura de formatos históricos en el producto.
4. Retirar assets gestionados antiguos sólo cuando su hash coincida con la versión conocida o exista una resolución explícita del conflicto.
5. Verificar que `install`, `sync`, `upgrade` y `verify` no reintroduzcan reglas o assets retirados.
6. Eliminar imports, ramas, tests y dependencias del modelo retirado, actualizando o eliminando todas las llamadas internas afectadas.
7. Verificar que no sobrevivan bridges, aliases, parsers antiguos ni adapters mantenidos sólo por retrocompatibilidad.

Un archivo modificado por el usuario se preserva y se informa como conflicto con recuperación. Un archivo ajeno al manifest no se borra. Una interrupción no convierte una migración parcial en exitosa. Un segundo intento no duplica bloques ni elimina trabajo nuevo.
