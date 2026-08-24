# Tasks: <change>

- [ ] 1. Analysis and constraints
  - [ ] 1.1 Revisar archivos existentes, dependencias, ownership y patrones locales.
  - [ ] 1.2 Confirmar alcance, non-goals, restricciones y entorno real.
  - [ ] 1.3 Identificar riesgos, decisiones pendientes y necesidad de escalar Lite -> Full.

- [ ] 2. Design alignment
  - [ ] 2.1 Para Full, completar arquitectura, componentes, datos, seguridad, operaciones y diagramas.
  - [ ] 2.2 Para Lite, mantener contexto compacto y agregar diagramas solo si reducen ambigüedad.
  - [ ] 2.3 Confirmar criterios WHEN/THEN observables antes de implementar.

- [ ] 3. Implementation
  - [ ] 3.1 Implementar por bloques coherentes, sin mezclar cambios no relacionados.
  - [ ] 3.2 Preservar contratos públicos, seguridad, datos y defaults salvo autorización explícita.
  - [ ] 3.3 Actualizar tests/docs ligados al cambio.

- [ ] 4. Validation
  - [ ] 4.1 Ejecutar validación proporcional real.
  - [ ] 4.2 Registrar comandos, resultados y limitaciones bajo `verification/<change>/`.
  - [ ] 4.3 Verificar que `change-overview.html` fue refrescado por el lifecycle.

- [ ] 5. Sync and delivery readiness
  - [ ] 5.1 Para Full, ejecutar sync cuando los deltas estén validados.
  - [ ] 5.2 Revisar estado de gates: implemented, validated, delivery_pending, delivered o closed.
  - [ ] 5.3 No archivar ni reportar cierre sin evidencia de validación, sync y delivery cuando aplique.
