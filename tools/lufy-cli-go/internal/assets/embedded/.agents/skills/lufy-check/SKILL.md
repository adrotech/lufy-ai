---
name: lufy-check
description: Verificar un resultado con comprobaciones pertinentes cuando el usuario pida validación o el trabajo requiera evidencia reproducible.
---

# Comprobar el resultado

Aplica [.lufy/contracts/evidence.md](../../../.lufy/contracts/evidence.md) y las restricciones del proyecto.

- Parte del resultado pedido y de lo que cambió. Para código comprueba comportamiento y regresiones; para texto, fidelidad y exactitud; para investigación, fuentes e inferencias; para datos, consultas y procedencia; para operaciones, estado posterior.
- El ejecutor puede escribir pruebas, implementar y revisar su trabajo. Usa TDD o revisión independiente cuando aporten valor o sean requisitos explícitos. No exijas otro rol para continuar.
- Ejecuta sólo comprobaciones autorizadas y disponibles. Agrupa suites completas; vuelve a comprobar la evidencia afectada por cambios posteriores.
- Conserva evidencia observada y distingue aprobado, fallido, inconcluso y no aplicable. No cuentes no aplicable como aprobado ni una recuperación lanzada como resultado observado.
- Explica defectos materiales y límites pendientes. Ningún score compensa un defecto crítico.
- Usa los comandos de evidencia, candidate o review sólo si la integración lo necesita. No crees un archivo JSON, un registro o un documento por obligación de esta skill.
