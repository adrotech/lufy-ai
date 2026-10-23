# Diseño y confianza

El ensayo usa `codex exec` no ephemeral para que la delegación pueda conservar el rollout padre, con user config ignorada, apps/plugins desactivados y sandbox read-only. Conserva autenticación y restricciones heredadas y elimina únicamente los cuatro handles de la sesión host, igual que el piloto existente. Nunca modifica configuración global.

Se copian instrucciones y hooks Lufy a un fixture nuevo fuera del checkout. Un observer local normaliza los callbacks oficiales SessionStart/SubagentStart/SubagentStop/SessionEnd/PostToolUse y captura sólo los IDs seudonimizados, clasificación de payload y fingerprint. El observer llama al lifecycle Lufy real para sus eventos soportados. El transporte no depende del formato de transcripts.

La confianza de hooks es explícita y temporal: sólo se permite el modo de automatización para definiciones del fixture revisadas y hasheadas antes del proceso. No relaja sandbox/aprobaciones ni crea confianza persistida. El resultado distingue esta condición del comportamiento normal, donde hooks nuevos/no confiables se omiten. Referencias: https://learn.chatgpt.com/docs/hooks y https://learn.chatgpt.com/docs/agent-configuration/subagents, consultadas el 2026-10-22; el esquema del binario local complementa la documentación.

El primer payload nativo es null a propósito; se exige una recuperación con el mismo ID antes de aceptar un Result Contract sintético previamente validado. Esto prueba transporte/recuperación, no calidad de una review real. El collector sólo clasifica como válido el JSON idéntico al envelope validado esperado. Cualquier variante desconocida del contrato de herramientas queda inconclusa.

El evaluador exige orden y vinculación de callbacks/herramientas, dos terminaciones del hijo, cierre reconocido y fin de sesión. Ausencias y fallos de infraestructura no son éxito; una violación observada conserva precedencia. Los artifacts del modelo quedan fuera del repositorio. Se acota cada intento; no se repite buscando verde sin nueva evidencia o diagnóstico.
