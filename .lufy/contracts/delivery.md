# Entrega autorizada

El agente principal conserva la responsabilidad del resultado. Commit, push, PR, promoción, release, publicación y cambios externos requieren autorización explícita del usuario. Una autorización vigente permite continuar dentro de su alcance sin nuevos checkpoints metodológicos.

Antes de entregar, inspecciona el diff y el estado Git, preserva cambios ajenos y comprueba el comportamiento afectado con herramientas disponibles. Registra comandos reales y sus resultados; no afirmes éxito por ejecutar una orden ni por un autorreporte.

En Lufy, `develop` es la base de integración y `main` la rama estable. No trabajes diariamente en ramas protegidas ni hagas force push sin pedido explícito. Los tags estables `v*` deben apuntar a commits alcanzables desde `main`.

Para un PR contra `develop`, comprueba whitespace con `git diff --check origin/develop...HEAD` sobre commits y `git diff --check origin/develop` si hay cambios pendientes. Revisa los checks remotos requeridos y observa su conclusión. Si falta evidencia o hay fallos, informa lo pendiente y su recuperación sin declarar la entrega completa.

La limpieza de ramas y archivos sólo procede cuando es segura y está dentro del alcance. No borres trabajo local o datos útiles para cerrar un proceso.
