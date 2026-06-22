## 1. OpenSpec

- [x] Crear propuesta y spec delta.
- [x] Validar OpenSpec strict.

## 2. Registry

- [x] Definir comandos soportados, parametros, choices y defaults.
- [x] Agregar tests de generacion de args.

## 3. TUI

- [x] Implementar pantalla de seleccion de comandos.
- [x] Implementar pantalla de parametros con bool/choice/text.
- [x] Permitir confirmar/cancelar y devolver args.

## 4. Integracion CLI

- [x] Abrir palette cuando `lufy-ai` se ejecuta sin args en TTY.
- [x] Mantener help para no-TTY.
- [x] Reutilizar `cli.Run` para ejecutar args generados.

## 5. Validacion

- [x] Ejecutar tests Go aplicables.
- [x] Ejecutar `openspec validate "add-cli-command-palette" --strict`.
- [x] Ejecutar `scripts/validate.sh`.
- [x] Verificar binario local si aplica.
