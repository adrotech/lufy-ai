# Diseño

Metrics usa un mapa de inicios por (run_id, task_ref). Un inicio duplicado vuelve ambiguo ese par y no sustituye silenciosamente su timestamp. Finalizaciones inválidas no suman duración. Rework/reopened con timestamp ausente conservan su count observado y marcan partial. Los nombres de los campos públicos se conservan.

Review slice único: service.go, nuevo metrics_integrity_test.go, reporte docs, delta/spec sincronizada y mirror. Los archivos de pruebas pertenecen a test-writer; implementación/docs a root. Validación agrupada después del join: pruebas Go, coverage/vet/build disponibles y CLI contra fixtures fuera del checkout. No se paralelizan archivos compartidos ni delivery.

La matriz de disponibilidad distingue contrato portable, productor lifecycle Codex y runtime observado. No se inicializa configuración, grafo o memoria del repo como efecto lateral. workflow_limits no disponible: sin budgets inventados. Overview OpenSpec: not_available.
