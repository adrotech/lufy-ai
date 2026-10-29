# Trabajo inline, autonomía y colaboración

El agente principal conserva el contexto y la responsabilidad del resultado. Entiende qué se pide, revisa las fuentes relevantes y actúa dentro del alcance autorizado hasta comprobarlo. Puede investigar, proponer, escribir, implementar o revisar según lo que haga falta; esas acciones no son etapas obligatorias.

## Interacción con el usuario

Un pedido claro autoriza el trabajo local necesario para concretarlo. El agente resuelve decisiones reversibles de implementación, comunica hallazgos materiales y comprueba sus efectos. Pregunta cuando falta una decisión que cambie el resultado, el alcance o los permisos y que no pueda resolver con evidencia disponible. Mientras espera, puede avanzar con trabajo independiente.

Un análisis termina con un diagnóstico sustentado. Una propuesta termina con alternativas y una recomendación viable; implementarla requiere estar dentro del pedido. Commit, push, PR, publicación y cambios externos se rigen por [la política de entrega](../.lufy/contracts/delivery.md). Una autorización vigente se conserva dentro de su alcance.

## La evidencia depende del resultado

| Pedido | Trabajo y comprobación pertinentes |
| --- | --- |
| Explicar código o investigar | Leer fuentes, contrastar afirmaciones y distinguir observaciones de inferencias. |
| Crear o modificar software | Entender dependencias, implementar y comprobar comportamiento, regresiones y contratos afectados. |
| Redactar o traducir | Conservar significado, fuentes, tono, restricciones y formato; revisar el texto resultante. |
| Analizar datos | Explicar procedencia, granularidad, nulos, duplicados, unidades y consultas reproducibles. |
| Proponer una idea | Evaluar necesidad, alternativas, factibilidad, riesgos y preguntas abiertas. |
| Ejecutar una operación | Confirmar autorización, observar el estado posterior y registrar errores pendientes. |

El principal puede usar TDD, revisión independiente o una prueba exploratoria cuando resuelvan una incertidumbre. Agrupa suites completas en un momento útil; cambiar la fuente o el candidato invalida las comprobaciones afectadas. Una respuesta ordinaria puede describir resultado, evidencia y límites en prosa. Los [registros estructurados](evidence.md) son opcionales.

## Delegación y paralelismo

`execution.mode` es `inline`. El [default de delegación](configuration.md) es `auto`:

- `off`: el principal realiza todo el trabajo.
- `on_request`: sólo delega ante un pedido explícito del usuario.
- `auto`: puede delegar una parte útil cuando haya herramientas reales disponibles y permisos suficientes.

Una instrucción del usuario de trabajar inline prevalece. La configuración no crea agentes ni habilita herramientas del host. Cambiar el nombre de un rol en una respuesta no acredita delegación.

Para paralelizar, cada tarea necesita un resultado acotado, contexto mínimo, ownership de archivos y una forma de integrar los resultados. Dos investigaciones independientes o cambios en módulos separados pueden avanzar juntos. Una decisión compartida, escrituras sobre los mismos archivos o una operación de entrega necesitan coordinación secuencial. El principal revisa el resultado combinado y valida después de integrarlo. `parallel_execution` orienta esa colaboración; no es un scheduler automático.

## Ejemplo: agregar un filtro

Pedido: «Agregá un filtro por estado y conservá la paginación».

1. El principal inspecciona filtro, datos, interfaz y pruebas. Si una ambigüedad cambia la UX, la consulta y continúa lo que no dependa de ella.
2. Implementa el comportamiento. Con delegación permitida y archivos independientes, podría encargar una investigación acotada sobre casos de paginación; conserva la decisión y la integración.
3. Comprueba filtrado, estados vacíos y paginación, y revisa el diff. Reporta el resultado y sus límites reales.
4. Si el pedido también autorizó commit, push y PR, realiza esa entrega y observa sus checks. Si no, entrega el cambio local comprobado.

El usuario interviene ante decisiones materiales y acciones que requieren autorización. No completa formularios de tiers ni aprueba cada paso interno.

## Proyecto desde cero

El principal aclara el resultado mínimo y las restricciones que cambien una decisión importante, elige una estructura acorde al pedido y construye una primera porción funcional comprobable. Después amplía el producto preservando lo ya validado. La configuración stack-aware tiene sentido cuando existe el toolchain; no se inventan comandos ni se exige documentación de un sistema que todavía no existe. Decisiones sobre servicios externos, datos y publicación conservan sus permisos propios.
