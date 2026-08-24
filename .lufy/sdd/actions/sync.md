# Sync

En Full, aplica los deltas validados a specs activas mediante `lufy-ai sdd sync --change <name>`. Bloquea ante targets ambiguos, artifacts inválidos o cualquier fallo de preflight. En Lite, reporta `not_applicable` sin mutaciones. Sync refresca el overview y no archiva.
