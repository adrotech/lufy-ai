# codex-runtime-diagnostics

## Purpose

Definir diagnóstico Codex explícito, acotado y sin inferencias de capacidades nativas.

## Requirements

### Requirement: Bounded explicit runtime diagnostics
El doctor SHALL permitir elegir explícitamente un ejecutable Codex y consultar sólo versión y flags con deadline total, salida acotada y códigos de causa sin contenido crudo. El reporte SHALL distinguir configuración estática, probes observados y delegación nativa no comprobada.

#### Scenario: Explicit runtime probe
- **WHEN** se solicita doctor con --codex y un ejecutable resoluble
- **THEN** consulta sólo --version y features list, sin shell ni modelos, con límite total de cinco segundos y 64 KiB por stdout

#### Scenario: Partial evidence is preserved
- **WHEN** un probe falla, supera el límite o devuelve evidencia inválida
- **THEN** el reporte conserva la evidencia independiente válida, devuelve códigos seguros y no infiere delegación nativa de las flags

#### Scenario: Existing doctor remains compatible
- **WHEN** no se solicita el diagnóstico de runtime
- **THEN** el doctor conserva su salida previa sin añadir un reporte runtime vacío

### Requirement: Explicit safe local activation
La activación de memoria y grafo SHALL preservar configuración y archivos existentes y SHALL distinguir opciones elegidas de límites previamente configurados.

#### Scenario: Local project activation
- **WHEN** el usuario autoriza activar memoria y grafo del repositorio
- **THEN** se utiliza configuración local sin cambiar modelos, credenciales ni permisos globales y se valida el estado resultante
