# Diseño

Programa T1; slice T2 experimental. Primero B2 y B3 establecen baseline disponible/partial/unavailable. Se preservan muestras históricas y hashes. El candidato sólo cambia la entrega de archivos sintéticos inline frente a referencias; las instrucciones de roles/contrato, sandbox y gates permanecen. La ejecución del candidato usa una muestra nueva, no reemplaza resultados fallidos.

Una unidad revisable cubre experimentación/reporte y pruebas pertinentes. Cambios compartidos en pilot.py dependen del cierre de su implementación B2; ejecución secuencial, sin mutaciones Git/GH paralelas. No se cambia configuración global ni se adopta una política en producción por una única muestra. Sin prueba reproducible, decisión keep_baseline.

Comparaciones adicionales: recuperación Lite/checkpoints e idempotencia se verifican sobre mecanismos existentes; aprobación reutilizable sólo dentro del alcance explícito, como la autorización global de esta sesión. TDD especializado frente a unidad de comportamiento se documenta como desk check cuando no exista un ensayo causal pareado; nunca se atribuye mejora de latencia/coste a esa evidencia. Este límite impide adoptar cambios de routing no demostrados.

Límites del experimento se declaran como tamaño del ensayo, no como workflow_limits del proyecto. La configuración de límites sigue not_available. Overview OpenSpec: not_available.
