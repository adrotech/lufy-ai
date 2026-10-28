# Ensayos del principal inline

- [WORK.md](WORK.md): fixtures y runner opt-in por familia.
- [FAMILIES.md](FAMILIES.md): auditoría y límites de la evidencia.
- `comparison.py`: comparación aislada de snapshots/instalaciones A/B/C y seguimientos. La variante histórica se proporciona mediante un ejecutable externo; no mantiene un motor de compatibilidad en el producto.

Listar casos con `python3 tools/codex-behavior-pilot/work.py` no invoca modelos ni escribe evidencia. Una ejecución real requiere `--live` y output externo nuevo; no se activa desde validación, hooks ni instalación. Conservar métricas ausentes como null y no declarar adopción automática.

Los ensayos anteriores ligados a roles y envelopes se retiraron de la suite actual. Su historia permanece en Git y en los respaldos privados, sin interpretar casos pendientes como aprobados. Se conserva la calibración del oráculo, aislamiento, trazas, alteraciones y límites del runner actual.
