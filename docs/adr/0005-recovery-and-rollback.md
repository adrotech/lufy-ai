# ADR 0005: Rollback acotado con backups existentes

## Estado

Aceptada.

## Contexto

Las operaciones `install` y `sync` pueden fallar después de haber escrito algunos archivos. Un diseño two-phase completo reduciría ese riesgo, pero aumenta complejidad y superficie de staging.

## Decisión

Cuando una operación ya creó un backup de recovery, un error posterior intenta rollback automático restaurando los archivos capturados en ese backup.

No se intenta rollback automático de instalaciones iniciales sin backup porque no hay estado anterior inequívoco para todos los archivos nuevos.

## Consecuencias

- Los errores parciales después de backup requieren menos recuperación manual.
- Si el rollback automático falla, el error conserva la ruta del backup para restauración manual.
- Un diseño two-phase puede evaluarse en el futuro si aparece una necesidad más fuerte.

## Generación inline

Los backups actuales usan schema 2 y declaran `runtimeStateSchema: 4`. Restore público rechaza otras generaciones y payload inválido antes de escribir; rollback interno permite conservar preimágenes dañadas. La recuperación entre generaciones usa un snapshot conjunto con el ejecutable correspondiente, fuera del runtime.
