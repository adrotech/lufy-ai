# Diseño

Servicio stdlib `internal/codexruntime` con resolver y runner inyectables. Un único contexto limita ambos procesos; stdout de cada probe tiene máximo 64 KiB y stderr se descarta. El runner usa exec.CommandContext y WaitDelay, sin prometer contención de todo árbol de procesos. El servicio no recibe credenciales ni lee su contenido.

Reporte aditivo `codex-runtime-report/v1` en doctor únicamente con flag explícita. `available` requiere versión y features válidas, `partial` una evidencia válida y `unavailable` ninguna. Flags duplicadas conocidas invalidan el probe completo. Se conserva el PATH resuelto como dato de comparación, sin llamar shadowing a una selección explícita distinta. El doctor estático conserva sus gates independientes de instalación.

La activación local usa los comandos públicos de memoria y contexto, con configuración mínima explícita sólo si falta. No modifica modelos, autenticación ni permisos globales. El estado local no es un asset distribuido ni evidencia de capacidades nativas.
