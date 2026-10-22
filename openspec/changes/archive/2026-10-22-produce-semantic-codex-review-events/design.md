# Diseño

`reviewevents.Service` recibe Store Run Ledger y reloj inyectables. Opciones: RunID, ReviewID, Action, OperationID, Result y EvidenceSHA256. ReviewID identifica un intento; OperationID distingue ocurrencias de rework/reopen y conserva idempotencia. Ambos son tokens opacos validados; sólo se persisten hashes. Start/complete tienen como máximo una ocurrencia por intento.

El servicio reutiliza LoadRun/Append con ExpectedLocalSequence y hasta tres intentos de CAS. Para retries conserva OccurredAt del evento ya durable; conflictos de contenido no se convierten en noop. Complete/rework/reopen requieren un inicio conocido; relojes anteriores al inicio se rechazan. Las acciones emiten KindEvidence y nunca Checkpoint ni KindFinish.

La CLI decodifica/canonicaliza el Result Contract y evalúa el claim reviewer mediante políticas existentes. El resultado passed no es compatible con blocked/escalated. El fingerprint sólo identifica ese contrato, no atesta identidad nativa ni candidato Git; el bloque 5 liga evidencia a candidato. El ledger guarda una referencia canónica hasheada y el outcome explícito, no paths ni cuerpos de review.

El host o usuario ejecuta el productor antes/después de la review real. Los roles read-only devuelven sus contratos; no adquieren permiso para escribir metadata ni ejecutar shell. El actor operativo no se infiere de hooks. Las métricas son observación, no autorización ni prueba de calidad.
