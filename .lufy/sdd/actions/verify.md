# Verify

Ejecuta validación strict —que también refresca `change-overview.html`—, contrasta requirements/scenarios con código y tests, y registra evidencia bajo `verification/<change>/`.

La verificación debe revisar:

- completeness: objetivos del LLM, restricciones, entorno, aceptación y tasks cubiertos;
- correctness: implementación alineada con requirements/scenarios y comportamiento esperado;
- coherence: diseño, patrones, diagramas, datos y seguridad consistentes con el código;
- proportionality: Lite no creció a Full sin escalar; Full cubre arquitectura y riesgos suficientes;
- evidence: comandos reales, resultados, limitaciones y revisión estática documentados.

Reporta por separado completeness, correctness, coherence y gate state.
