# Política de seguridad

## Reportar vulnerabilidades

Reporta vulnerabilidades de forma privada usando GitHub Security Advisories del repositorio o contactando a los mantenedores antes de abrir un issue público.

Incluye, si es posible:

- Versión afectada o commit.
- Sistema operativo y shell.
- Pasos de reproducción.
- Impacto esperado.
- Logs o salida relevante sin secretos.

## Alcance

Están dentro de alcance:

- Path traversal o escrituras fuera del target permitido.
- Corrupción de archivos gestionados por install/sync/restore.
- Problemas de supply chain en release artifacts, checksums o workflows.
- Ejecución inesperada de comandos durante bootstrap o instalación.

No incluyas secretos reales en reportes, fixtures o pruebas.

## Expectativas

El proyecto prioriza fixes de seguridad sobre mejoras funcionales. Los detalles públicos se coordinarán después de tener mitigación o release disponible.
