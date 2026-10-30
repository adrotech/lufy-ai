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
- Corrupción de archivos gestionados por install/sync/restore, backups o validación de generación.
- Problemas de supply chain en release artifacts, checksums o workflows.
- Ejecución inesperada de comandos durante instalación, upgrade o lecturas Git del candidato.

No incluyas secretos reales en reportes, fixtures o pruebas.

## Expectativas

El proyecto prioriza fixes de seguridad sobre mejoras funcionales. Los detalles públicos se coordinarán después de tener mitigación o release disponible.

La [seguridad de releases](docs/release-security.md) distingue checksum, contrato de formato y firma. Las comprobaciones declaradas y los registros locales no autentican un agente ni sustituyen los permisos del host. Los ensayos con modelos usan autorización explícita y no corren como parte de la validación offline.
