# Documentación de Lufy

Estas guías describen el flujo inline integrado en `develop`. La disponibilidad del código y la publicación de una release son estados distintos: ver [estado actual](status.md).

## Usar Lufy

| Necesidad | Guía |
| --- | --- |
| Aprender con ejemplos, conversaciones y diagramas interactivos | [Manual visual](manual/README.md) |
| Entender autonomía, preguntas y paralelismo | [Flujo de trabajo](harness-workflow.md) |
| Instalar, sincronizar o recuperar | [Instalación](installation.md) |
| Ajustar delegación, superficies y capacidades | [Configuración](configuration.md) |
| Consultar operaciones y flags | [Comandos](cli.md) |
| Buscar memoria o relaciones entre archivos | [Memoria y contexto](memory-context.md) |
| Inspeccionar comprobaciones y registrar reviews | [Evidencia opcional](evidence.md) |
| Consultar ejecuciones explícitas | [Run Ledger](run-ledger.md) |
| Resolver fallos o drift | [Diagnóstico](troubleshooting.md) |

## Mantener el producto

- [Arquitectura](architecture.md), [decisiones ADR](adr/README.md) y [aprendizajes](lessons/lufy-ai.md).
- [Contribuciones](../CONTRIBUTING.md), [CLI Go](../tools/lufy-cli-go/README.md) y [validación local/CI](ci-quality-gates.md).
- [Preparación de releases](releasing.md), [seguridad de artifacts](release-security.md), [ramas GitHub](github-branch-settings.md) y [reporte de vulnerabilidades](../SECURITY.md).
- [Estado](status.md), [hipótesis de evolución](roadmap.md) e [historia de investigación e implementación](history/README.md).

## Fuentes y mantenimiento

El comportamiento ejecutable se comprueba contra `tools/lufy-cli-go` y el help de su binario. Las políticas distribuidas son [trabajo](../.lufy/contracts/work.md), [evidencia](../.lufy/contracts/evidence.md) y [entrega](../.lufy/contracts/delivery.md). La configuración particular del usuario no es un default del producto.

Los assets editoriales seleccionados se mantienen en la raíz y se copian byte a byte a `tools/lufy-cli-go/internal/assets/embedded`. Las guías de `docs/` no son assets instalables. Cambiar un mirror sin su fuente deja una distribución inconsistente.

Los informes fechados conservan su evidencia y condiciones originales. Sus comandos, permisos, roles o pendientes no se heredan como instrucciones actuales. El changelog conserva las releases anteriores y reserva `Unreleased` para cambios integrados todavía sin publicar.
