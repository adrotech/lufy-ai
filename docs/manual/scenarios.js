/* Recorridos didácticos. Estos datos no son trazas ni resultados de ejecuciones. */
window.LufyManualScenarios = [
  {
    "id": "analysis",
    "label": "Comprender",
    "title": "Entender antes de cambiar.",
    "kicker": "Análisis de código",
    "summary": "Explicar un comportamiento puede ser el resultado completo del trabajo.",
    "scope": "Lectura y explicación",
    "prompt": "Explicá cómo funciona la autenticación de este proyecto y dónde se controla la expiración. Quiero entenderlo; conservá los archivos como están.",
    "boundary": "Pedir una explicación no autoriza implementar una mejora descubierta durante el análisis.",
    "steps": [
      {
        "title": "Ubicar el contexto",
        "context": "Una actualización de progreso",
        "message": "Voy a seguir el recorrido de la autenticación desde la entrada hasta la validación de la sesión y localizar el control de expiración.",
        "action": "Lee las instrucciones y localiza entradas, middleware, servicios y pruebas disponibles.",
        "evidence": "Archivos y referencias concretas del proyecto, sin asumir un framework.",
        "skill": "lufy-mem-search puede aportar una decisión previa si existe memoria pertinente; se contrasta con el código."
      },
      {
        "title": "Seguir el recorrido",
        "context": "Trabajo autónomo",
        "message": "Estoy contrastando dónde se crea la sesión, dónde se valida y qué comportamiento ocurre al vencer. Voy a separar lo que veo en el código de lo que requeriría ejecutarlo.",
        "action": "Relaciona llamadas, configuración y dependencias; distingue ruta normal y vencimiento.",
        "evidence": "La conexión observable entre los puntos de entrada y los controles.",
        "skill": "El contexto lexical puede ayudar a localizar relaciones; el principal confirma sus fuentes."
      },
      {
        "title": "Contrastar",
        "context": "Una comprobación pertinente",
        "message": "La explicación va a señalar el control localizado y qué cubren las pruebas existentes. Si no puedo ejecutar una comprobación, lo voy a dejar explícito.",
        "action": "Revisa pruebas y configuración relevantes; ejecuta una comprobación no mutante sólo si es útil y está disponible.",
        "evidence": "Coherencia entre explicación, implementación y observaciones disponibles.",
        "skill": "lufy-check orienta la evidencia; no convierte la tarea en un cambio de código."
      },
      {
        "title": "Explicar",
        "context": "Cierre del pedido",
        "message": "Te dejo un mapa del recorrido, las referencias para seguirlo y los límites del análisis. Las posibles mejoras quedan como observaciones separadas.",
        "action": "Redacta una explicación navegable con referencias y límites.",
        "evidence": "Que la respuesta resuelva cómo funciona y dónde ocurre la expiración.",
        "skill": "Una respuesta útil puede terminar en prosa, sin crear artifacts adicionales."
      }
    ],
    "final": {
      "message": "El recorrido pasa por entrada, validación de sesión y acceso al recurso. En este ejemplo, la expiración se comprueba durante la validación. La explicación enlaza cada punto con su fuente.\nLa revisión estática permite seguir el flujo; el comportamiento del proveedor externo queda sin verificar si no hubo acceso.",
      "artifact": "Una explicación con mapa, referencias y observaciones.",
      "acceptance": "Podés localizar la expiración y seguir el recorrido desde la respuesta.",
      "boundary": "Los archivos se conservan y las mejoras no se implementan como parte de este pedido."
    }
  },
  {
    "id": "bug",
    "label": "Corregir",
    "title": "Encontrar la causa y corregirla.",
    "kicker": "Corrección de un bug",
    "summary": "El principal lleva la investigación hasta una corrección y sus comprobaciones.",
    "scope": "Cambios locales y comprobación",
    "prompt": "El listado queda vacío al filtrar si estaba en otra página. Investigá la causa, corregila y comprobá filtros y paginación. Conservá el contrato de la API.",
    "boundary": "Analizar y corregir cubre la unidad local pedida. La publicación y la entrega Git requieren su autorización.",
    "steps": [
      {
        "title": "Reproducir",
        "context": "Una actualización de progreso",
        "message": "Voy a revisar cómo se combinan filtro y página y buscar una reproducción que distinga el fallo del comportamiento esperado.",
        "action": "Inspecciona el estado, el flujo de datos y las pruebas del listado.",
        "evidence": "Un caso reproducible: cambiar el filtro desde una página que deja de existir.",
        "skill": "El principal puede investigar y escribir una prueba en la misma unidad de trabajo."
      },
      {
        "title": "Aislar la causa",
        "context": "Un hallazgo de ejemplo",
        "message": "La hipótesis de este caso es que se conserva una página fuera de rango al cambiar el filtro. Voy a comprobarla antes de ajustar el comportamiento.",
        "action": "Contrasta la hipótesis con el código y un caso discriminante.",
        "evidence": "La relación causal entre estado de página, filtro y lista vacía.",
        "skill": "TDD puede ayudar a fijar el comportamiento; no exige un subagente."
      },
      {
        "title": "Corregir",
        "context": "Trabajo autónomo",
        "message": "Voy a ajustar la transición del filtro para conservar una página válida. Mantengo el contrato existente y los cambios dentro del listado.",
        "action": "Implementa una corrección acotada sin modificar contratos externos.",
        "evidence": "Que el ajuste ataque la causa y conserve los casos válidos.",
        "skill": "No hace falta consultar nombres o detalles convencionales que el contexto permite resolver."
      },
      {
        "title": "Comprobar",
        "context": "Validación del comportamiento",
        "message": "Voy a comprobar el cambio de filtro, la navegación entre páginas y el resultado vacío legítimo, además de la validación disponible del módulo.",
        "action": "Ejecuta las pruebas relevantes y revisa el diff; agrupa la suite cuando corresponde.",
        "evidence": "Resultados reales de las comprobaciones y regresiones pertinentes.",
        "skill": "lufy-check ayuda a distinguir una observación aprobada de una comprobación pendiente."
      },
      {
        "title": "Cerrar",
        "context": "Resultado y límites",
        "message": "El cierre va a indicar la causa, el cambio, las comprobaciones observadas y cualquier límite de verificación. La entrega Git conserva su autorización propia.",
        "action": "Relaciona el resultado con su evidencia y revisa si falta algo del pedido.",
        "evidence": "Que el listado filtre y pagine de acuerdo con la aceptación.",
        "skill": "Si aparece un aprendizaje reusable, puede capturarse en memoria; no se registra cada turno."
      }
    ],
    "final": {
      "message": "En este ejemplo se corrigió la página fuera de rango al cambiar el filtro. La respuesta final detalla las comprobaciones de filtrado, paginación y vacío legítimo y sus resultados observados.\nEl contrato de la API se conserva. Cualquier comprobación que no pudo ejecutarse queda indicada.",
      "artifact": "Corrección local, pruebas pertinentes y explicación de causa y efecto.",
      "acceptance": "La reproducción deja de fallar y las regresiones relevantes se comprueban.",
      "boundary": "No se crea commit, PR o publicación salvo que estén autorizados."
    }
  },
  {
    "id": "feature",
    "label": "Agregar algo",
    "title": "Una funcionalidad con decisiones visibles.",
    "kicker": "Nueva funcionalidad",
    "summary": "El agente consulta una decisión de producto cuando cambia lo que va a construir.",
    "scope": "Funcionalidad local",
    "prompt": "Agregá notificaciones para los cambios de estado de los pedidos. Reutilizá la estructura actual y comprobá el flujo completo.",
    "boundary": "Una pregunta material no bloquea el trabajo independiente. El silencio no autoriza la parte pendiente.",
    "steps": [
      {
        "title": "Entender",
        "context": "Una actualización de progreso",
        "message": "Voy a revisar los cambios de estado, las superficies conectadas y los mecanismos de notificación existentes para definir una unidad integrada.",
        "action": "Inspecciona arquitectura, dependencias y restricciones del proyecto.",
        "evidence": "Dónde se origina el evento y dónde debe observarse la notificación.",
        "skill": "El perfil de superficies y plan pueden ayudar a identificar contratos; no ejecutan validaciones."
      },
      {
        "title": "Resolver una decisión",
        "context": "Una consulta necesaria",
        "message": "¿Las notificaciones deben aparecer sólo dentro de la app o también enviarse por email? Recomiendo empezar dentro de la app si ése cubre el uso inicial. Mientras lo definís, puedo revisar el evento y sus pruebas.",
        "reply": "Por ahora sólo dentro de la app. No agreguemos email.",
        "action": "Consulta el canal, que cambia UX, dependencias y alcance, y continúa la investigación independiente.",
        "evidence": "Una decisión explícita antes de construir la parte que depende del canal.",
        "skill": "El usuario decide producto; el principal resuelve los detalles técnicos dentro de ese alcance."
      },
      {
        "title": "Construir",
        "context": "Trabajo autónomo",
        "message": "Con el canal definido, voy a conectar el cambio de estado con la notificación dentro de la app y respetar la estructura del proyecto.",
        "action": "Implementa una unidad coherente. Con delegación permitida, módulos independientes podrían trabajarse por separado.",
        "evidence": "Contrato compartido resuelto y ownership claro antes de un posible paralelo.",
        "skill": "La delegación es opcional. Usar una skill no equivale a repartir la implementación."
      },
      {
        "title": "Integrar",
        "context": "Comprobación del conjunto",
        "message": "Voy a comprobar la notificación desde el cambio de estado hasta su visualización, incluyendo duplicados y estados vacíos relevantes.",
        "action": "Reúne los cambios, revisa su interacción y valida el recorrido completo.",
        "evidence": "El comportamiento integrado, además de las comprobaciones de cada módulo.",
        "skill": "lufy-check orienta regresiones y contratos; el autorreporte de un trabajador no reemplaza observaciones."
      },
      {
        "title": "Entregar resultado",
        "context": "Cierre del pedido",
        "message": "El resultado va a detallar el canal implementado, las comprobaciones y lo que quedó fuera del alcance acordado.",
        "action": "Explica el resultado y los límites vigentes.",
        "evidence": "Que la notificación corresponda al evento y al canal decidido.",
        "skill": "Una decisión durable sobre canales podría documentarse con lufy-mem-document si el proyecto usa memoria."
      }
    ],
    "final": {
      "message": "En este recorrido, las notificaciones se muestran dentro de la app al cambiar el estado de un pedido. El cierre describe los archivos relevantes y las comprobaciones del flujo integrado.\nEmail queda fuera del alcance acordado. Las limitaciones de verificación se informan explícitamente.",
      "artifact": "Una funcionalidad integrada y su explicación de uso.",
      "acceptance": "El cambio de estado produce la notificación según el canal acordado.",
      "boundary": "La decisión sobre email no se toma por ausencia de respuesta ni por iniciativa del agente."
    }
  },
  {
    "id": "app",
    "label": "Crear de cero",
    "title": "De una idea a una primera porción funcional.",
    "kicker": "Proyecto desde cero",
    "summary": "Primero se concreta un uso comprobable; después se amplía preservando lo que funciona.",
    "scope": "Proyecto local nuevo",
    "prompt": "Creá una app personal para organizar tareas. Quiero empezar con una versión local, con creación, edición y filtros, que pueda probar en el navegador.",
    "boundary": "Crear la app local no autoriza contratar servicios, cargar datos reales o publicarla.",
    "steps": [
      {
        "title": "Concretar el uso",
        "context": "Una actualización de progreso",
        "message": "Voy a construir una primera versión local centrada en crear, editar y filtrar tareas. Revisaré el entorno antes de elegir una estructura y comandos de comprobación.",
        "action": "Define aceptación observable y detecta herramientas y restricciones disponibles.",
        "evidence": "Un recorrido concreto que permita probar el uso pedido.",
        "skill": "El principal puede elegir detalles convencionales; consulta sólo decisiones que cambien materialmente el producto."
      },
      {
        "title": "Dar una base",
        "context": "Trabajo autónomo",
        "message": "Voy a preparar una estructura acorde al proyecto y conectar la primera acción de usuario con su estado y visualización.",
        "action": "Crea una base mínima y una primera porción integrada.",
        "evidence": "Que la app arranque y la primera acción tenga efecto visible.",
        "skill": "La configuración stack-aware tiene sentido cuando existe el toolchain; no se inventan scripts."
      },
      {
        "title": "Completar el recorrido",
        "context": "Construcción incremental",
        "message": "Voy a incorporar edición y filtros sobre la misma base, conservando el comportamiento de creación.",
        "action": "Amplía la porción funcional y mantiene coherencia entre estado, datos e interfaz.",
        "evidence": "Que las acciones se combinen sin romper las anteriores.",
        "skill": "Los límites de estructura del usuario son aceptación; un build por sí solo no comprueba esa estructura."
      },
      {
        "title": "Probar como usuario",
        "context": "Verificación útil",
        "message": "Voy a probar el recorrido en navegador y las comprobaciones disponibles, incluyendo estados vacíos y filtros que no encuentran resultados.",
        "action": "Revisa comportamiento e integración, con evidencia visual cuando aporta.",
        "evidence": "Acciones observadas y resultados reales, no sólo que el servidor arrancó.",
        "skill": "lufy-check permite elegir evidencia pertinente a la app y sus restricciones."
      },
      {
        "title": "Dejarla utilizable",
        "context": "Cierre de la primera versión",
        "message": "Te voy a dejar cómo abrirla, qué hace esta versión y qué decisiones quedan para ampliar el producto.",
        "action": "Explica uso, alcance y limitaciones de la entrega local.",
        "evidence": "Que el usuario pueda probar la primera versión y entender sus límites.",
        "skill": "Una próxima propuesta puede partir de este resultado sin rehacerlo por una etapa metodológica."
      }
    ],
    "final": {
      "message": "La primera versión del ejemplo permite crear, editar y filtrar tareas localmente. El cierre incluye instrucciones para abrirla y las comprobaciones realmente observadas.\nCuenta compartida, sincronización y publicación quedan como decisiones de una ampliación futura.",
      "artifact": "Una primera versión funcional local e instrucciones de uso.",
      "acceptance": "El recorrido pedido se puede probar y las acciones conservan su integración.",
      "boundary": "No se confunde una demo local con una operación productiva o una publicación."
    }
  },
  {
    "id": "proposal",
    "label": "Pensar una idea",
    "title": "Una propuesta puede ser el destino.",
    "kicker": "Propuesta y diseño",
    "summary": "El trabajo termina con una decisión mejor informada, sin implementar por inercia.",
    "scope": "Análisis y propuesta",
    "prompt": "Proponé cómo incorporar búsqueda al producto. Compará dos alternativas viables según lo que ya tenemos y recomendá una. Todavía no lo implementes.",
    "boundary": "Proponer una dirección no autoriza agregar dependencias o cambiar el producto.",
    "steps": [
      {
        "title": "Leer la necesidad",
        "context": "Una actualización de progreso",
        "message": "Voy a revisar qué contenido se necesita encontrar, el volumen y la infraestructura actual antes de comparar alternativas.",
        "action": "Relaciona la necesidad con capacidades y restricciones existentes.",
        "evidence": "Fuentes del proyecto y supuestos explícitos, sin inventar volumen o costos.",
        "skill": "La memoria puede aportar decisiones anteriores; el estado actual conserva precedencia."
      },
      {
        "title": "Comparar opciones",
        "context": "Investigación útil",
        "message": "Voy a comparar una opción simple sobre el almacenamiento actual y otra con índice dedicado, según complejidad, operación y comportamiento esperado.",
        "action": "Investiga alternativas y verifica información cambiante cuando corresponde.",
        "evidence": "Factibilidad, diferencias relevantes y fuentes pertinentes.",
        "skill": "Una investigación distingue hechos, inferencias y aspectos todavía no medidos."
      },
      {
        "title": "Poner a prueba la recomendación",
        "context": "Contraste",
        "message": "Voy a revisar qué supuesto haría cambiar la recomendación y qué necesitaríamos medir para decidir con más confianza.",
        "action": "Busca límites y contraejemplos; propone una prueba acotada si resolvería incertidumbre.",
        "evidence": "Que la recomendación dependa de criterios identificables, no de preferencias sin fundamento.",
        "skill": "lufy-check puede orientar la comprobación de coherencia de la propuesta."
      },
      {
        "title": "Proponer",
        "context": "Resultado del pedido",
        "message": "La propuesta va a resumir opciones, recomendación, tradeoffs y preguntas pendientes. La implementación queda para un pedido que la cubra.",
        "action": "Redacta una propuesta proporcionada y útil para decidir.",
        "evidence": "Una recomendación que conecte necesidad, restricciones y alternativas.",
        "skill": "No se necesita un documento persistido si el formato solicitado se resuelve en la respuesta."
      }
    ],
    "final": {
      "message": "La propuesta del ejemplo recomienda empezar con la alternativa más simple que cubre la necesidad actual, explica sus límites y cuándo convendría un índice dedicado.\nIncluye supuestos por verificar y una forma de evaluar la decisión. La implementación no forma parte de esta entrega.",
      "artifact": "Comparación y recomendación con fuentes y preguntas abiertas.",
      "acceptance": "Podés decidir el siguiente paso y entender qué haría cambiar la recomendación.",
      "boundary": "No se agregan dependencias, cuentas ni cambios de código como consecuencia automática del análisis."
    }
  },
  {
    "id": "writing",
    "label": "Escribir",
    "title": "Un texto fiel también se comprueba.",
    "kicker": "Redacción y traducción",
    "summary": "La calidad se observa en el significado, las fuentes y el formato solicitado.",
    "scope": "Documento y revisión editorial",
    "prompt": "Convertí estas notas en una guía para usuarios nuevos y prepará una versión en inglés. Conservá los nombres del producto, el significado y los ejemplos importantes.",
    "boundary": "Un build no prueba fidelidad editorial. Publicar o enviar el texto requiere estar dentro de la autorización.",
    "steps": [
      {
        "title": "Leer las fuentes",
        "context": "Una actualización de progreso",
        "message": "Voy a identificar el público, las ideas y las restricciones de las notas antes de ordenar la guía y su traducción.",
        "action": "Lee el material y marca términos, hechos y ejemplos que deben conservarse.",
        "evidence": "Procedencia y restricciones de lo que se puede afirmar.",
        "skill": "Las fuentes del usuario prevalecen sobre completar huecos con afirmaciones inventadas."
      },
      {
        "title": "Dar estructura",
        "context": "Trabajo editorial",
        "message": "Voy a ordenar la guía alrededor del primer uso y las dudas frecuentes, conservando los ejemplos que ayudan a comprender.",
        "action": "Redacta una estructura para el público indicado y mantiene la identidad del material.",
        "evidence": "Que la reorganización conserve información útil y referencias.",
        "skill": "La brevedad por sí sola no demuestra una mejor documentación."
      },
      {
        "title": "Traducir y contrastar",
        "context": "Fidelidad",
        "message": "Voy a contrastar ambas versiones por sentido, nombres, números, advertencias y ejemplos. Los términos del producto se van a conservar.",
        "action": "Traduce y compara con la fuente, en lugar de comprobar sólo que el texto suene natural.",
        "evidence": "Ausencia de omisiones y cambios de significado relevantes.",
        "skill": "lufy-check orienta fidelidad, exactitud y formato; no impone pruebas de código."
      },
      {
        "title": "Entregar el texto",
        "context": "Cierre editorial",
        "message": "El resultado va a incluir ambas versiones, el criterio de vocabulario y cualquier ambigüedad de las notas que siga abierta.",
        "action": "Revisa la versión final y explica límites que importan al lector.",
        "evidence": "Que los documentos cumplan público, contenido, tono y formato.",
        "skill": "Una revisión independiente puede ayudar cuando el riesgo o el pedido la requieren y la delegación esté permitida."
      }
    ],
    "final": {
      "message": "El ejemplo termina con una guía ordenada para el primer uso y su versión en inglés. Conserva nombres del producto y ejemplos, y señala las ambigüedades que no se pudieron resolver con las notas.\nLa respuesta explica qué se contrastó entre las versiones.",
      "artifact": "Guía y traducción en el formato pedido.",
      "acceptance": "El contenido es fiel, coherente, utilizable y está revisado contra su fuente.",
      "boundary": "No se convierte la redacción en una publicación ni se atribuye validación lingüística a un build."
    }
  },
  {
    "id": "data",
    "label": "Analizar datos",
    "title": "Explicar el total, sin tocar los datos.",
    "kicker": "Diagnóstico de datos",
    "summary": "Un join puede multiplicar filas: la evidencia debe explicar granularidad y agregación.",
    "scope": "Sólo lectura",
    "prompt": "Con estos datos de ejemplo, explicá por qué el total de ventas cambia de 300 a 400 al unir pagos. Quiero un diagnóstico reproducible, sólo lectura.",
    "boundary": "Analizar datos no autoriza borrar duplicados, cambiar consultas productivas ni migrar una base.",
    "steps": [
      {
        "title": "Entender el grano",
        "context": "Una actualización de progreso",
        "message": "Voy a comparar qué representa una fila en ventas y en pagos antes de interpretar el total como un duplicado.",
        "action": "Revisa procedencia, claves, unidades, nulos y relación entre tablas o archivos.",
        "evidence": "La granularidad de ambas fuentes y sus claves de unión.",
        "skill": "El principal puede inspeccionar CSV o consultas read-only disponibles; no asume acceso a una base real."
      },
      {
        "title": "Reproducir el caso",
        "context": "Datos ficticios del ejemplo",
        "message": "El fixture tiene una venta A de 100 con dos pagos y una venta B de 200 con uno. Después del join aparecen las ventas 100, 100 y 200.",
        "action": "Construye o inspecciona un caso mínimo que haga visible la multiplicación.",
        "evidence": "Dos ventas suman 300; tres filas del join suman 400 si se agrega el importe de venta.",
        "skill": "La reproducción conserva los datos originales y separa filas del join de entidades de venta."
      },
      {
        "title": "Contrastar la explicación",
        "context": "Comprobación reproducible",
        "message": "Voy a contrastar cantidades y sumas antes y después del join. La corrección conceptual depende de qué métrica se quiera calcular.",
        "action": "Compara conteos, claves y agregados; explica qué hipótesis resiste el caso.",
        "evidence": "Una relación causal reproducible, con el significado de la métrica explícito.",
        "skill": "lufy-check orienta procedencia, unidades y consultas reproducibles, no cobertura de una app."
      },
      {
        "title": "Explicar el diagnóstico",
        "context": "Resultado del análisis",
        "message": "La respuesta va a explicar el cambio de granularidad y cómo calcular la métrica correcta, sin aplicar cambios a datos o sistemas.",
        "action": "Entrega causa, reproducción, alternativas y límites del diagnóstico.",
        "evidence": "Que el resultado explique el salto de 300 a 400 con las filas del fixture.",
        "skill": "Una propuesta de corrección no se aplica automáticamente a una base productiva."
      }
    ],
    "final": {
      "message": "En el fixture, A aparece dos veces porque tiene dos pagos. Sumar el importe de venta después del join cuenta 100 dos veces: 100 + 100 + 200 = 400. Las dos ventas originales suman 300.\nLa solución debe conservar la granularidad de la métrica: por ejemplo, agregar pagos antes de unir si se necesita una fila por venta.",
      "artifact": "Diagnóstico con caso mínimo, conteos, sumas y explicación de la métrica.",
      "acceptance": "La causa se puede reproducir y distinguir de un duplicado real de la entidad.",
      "boundary": "No se modifican datos ni se da por verificado un sistema externo que no fue observado."
    }
  },
  {
    "id": "operation",
    "label": "Operar",
    "title": "Recuperar y observar el estado posterior.",
    "kicker": "Operación y recuperación",
    "summary": "Una operación termina cuando se observa su efecto, no cuando se lanza un comando.",
    "scope": "Restauración local autorizada",
    "prompt": "En este proyecto de prueba, restaurá el backup actual que te indico. Revisá primero el preview, conservá los archivos ajenos y verificá el estado posterior.",
    "boundary": "El ejemplo supone autorización explícita para ese target y backup. Un formato incompatible o trabajo posterior no se ignoran para forzar el restore.",
    "steps": [
      {
        "title": "Identificar el estado",
        "context": "Una actualización de progreso",
        "message": "Voy a identificar el ejecutable, el target y el backup señalado antes de planificar la restauración.",
        "action": "Observa la generación y el estado existente; confirma que corresponden al alcance recibido.",
        "evidence": "Identidad del target, backup y formato compatible.",
        "skill": "No se cambian números de schema para que un dato incompatible parezca válido."
      },
      {
        "title": "Revisar el preview",
        "context": "Antes de escribir",
        "message": "Voy a revisar qué archivos cambiaría el restore y qué contenido debe conservarse. Si aparece un conflicto o una generación distinta, me detengo en ese punto.",
        "action": "Ejecuta el preview autorizado y contrasta paths, hashes y cambios posteriores relevantes.",
        "evidence": "Un plan aplicable que preserve los límites y datos útiles.",
        "skill": "Las restricciones de generación y recuperación conservan precedencia sobre completar el procedimiento."
      },
      {
        "title": "Aplicar",
        "context": "Operación autorizada",
        "message": "Si el preview es válido y corresponde a lo autorizado, voy a aplicar la restauración y conservar su resultado para verificarlo.",
        "action": "Aplica la operación prevista sin ampliar target o borrar trabajo para conseguir una aceptación.",
        "evidence": "Salida real del comando y cualquier error o recuperación pendiente.",
        "skill": "Una autorización vigente no se vuelve a pedir por una etapa metodológica; no cubre cambios nuevos de alcance."
      },
      {
        "title": "Observar el resultado",
        "context": "Verificación posterior",
        "message": "Voy a comprobar el estado y los archivos resultantes. Si la operación falla o queda parcial, lo voy a informar con su recuperación concreta.",
        "action": "Inspecciona el estado posterior y las comprobaciones pertinentes.",
        "evidence": "Que el contenido esperado se recuperó y se conservaron los límites.",
        "skill": "lufy-check distingue un comando lanzado de un efecto observado."
      },
      {
        "title": "Cerrar o informar el bloqueo",
        "context": "Resultado honesto",
        "message": "El cierre va a indicar qué se restauró y cómo se comprobó, o el bloqueo exacto si no fue posible completar una recuperación válida.",
        "action": "Reporta éxito observado, parcialidad o bloqueo sin compensar un fallo con otros checks.",
        "evidence": "Estado posterior y errores pendientes, con su recovery si corresponde.",
        "skill": "Un aprendizaje de recuperación puede conservarse en memoria si es durable y la capacidad está disponible."
      }
    ],
    "final": {
      "message": "Si la restauración se completó, el ejemplo termina identificando el backup, los archivos restaurados y las comprobaciones del estado posterior.\nSi el preview detecta una generación incompatible o un conflicto que impide aplicar, el resultado es un bloqueo concreto; no un restore exitoso.",
      "artifact": "Estado recuperado y evidencia, o diagnóstico de bloqueo con recuperación propuesta.",
      "acceptance": "El efecto solicitado está observado y los errores pendientes se conservan visibles.",
      "boundary": "No se afirma éxito por ejecutar un comando ni se amplía la autorización a otro target."
    }
  }
];
