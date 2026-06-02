## 1. Preparacion y baseline

- [x] Confirmar baseline con `scripts/validate.sh`.
- [x] Registrar cobertura agregada y paquetes con menor cobertura relevante.
- [x] Identificar funciones/paquetes objetivo iniciales: `installer`, `syncer`, `verify`, `projectconfig`.
- [x] Definir criterio de no-regresion para el preset `opencode` + `openspec`.

## 2. Puertos de aplicacion

- [ ] Introducir puertos minimos para filesystem/state/catalog/runtime/clock solo donde el slice los use.
- [ ] Ensamblar implementaciones concretas en constructors/factories sin cambiar la CLI publica.
- [ ] Actualizar tests para usar fakes o temp dirs segun corresponda.
- [ ] Verificar que dominio y puertos neutrales no importen detalles concretos de runtime.

## 3. Separacion de servicios grandes

- [x] Extraer tipos/helpers de acciones de `installer` a componente revisable.
- [x] Extraer presentacion del plan de `installer` fuera de `Run`.
- [x] Extraer tipos/helpers de acciones de `syncer` a componente revisable.
- [x] Extraer presentacion del plan de `syncer` fuera de `Run`.
- [x] Extraer emision/presentacion de reportes de `verify` a `reportEmitter`.
- [ ] Extraer `installer` plan/apply completos a componentes revisables.
- [ ] Extraer `syncer` plan/apply completos y consolidar reglas compartidas reales con `installer`.
- [ ] Extraer `verify` check building/check running completos.
- [ ] Revisar `projectconfig` para separar scanning, merge y persistencia.

## 4. Acciones tipadas y strategy

- [x] Declarar tipos/constantes para acciones y estados que afecten comportamiento.
- [x] Reemplazar comparaciones de strings por tipos donde sea seguro.
- [x] Mantener dispatch actual sin strategy adicional porque el switch sigue siendo local y testeado.
- [x] Agregar tests de accion desconocida para `installer` y `syncer`.
- [x] Preservar tests existentes de confirmacion, backup/recovery y no-mutacion.

## 5. SOLID, clean code y reviewer gates

- [x] Documentar en el resultado del slice como se cumple SRP/OCP/LSP/ISP/DIP o que queda pendiente.
- [x] Reducir mezcla de decision y reporting en `installer`, `syncer` y `verify` para el primer slice.
- [ ] Eliminar duplicacion significativa entre install/sync sin crear abstracciones prematuras.
- [x] Validar que errores nuevos sigan siendo accionables y en espanol.

## 6. TDD / AAA

- [x] Para cada cambio de comportamiento del slice inicial, registrar RED/GREEN o `not_applicable` con razon.
- [x] Estructurar tests nuevos con AAA explicito o implicitamente claro.
- [ ] Agregar triangulacion para edge cases de path safety, drift, rollback, dry-run y mismatches de tool/metodologia.
- [x] Mantener tests de integracion end-to-end para compatibilidad observable.

## 7. Validacion final

- [x] Ejecutar `scripts/validate.sh`.
- [x] Ejecutar `openspec validate "strengthen-go-cli-hexagonal-boundaries" --strict` cuando el CLI este disponible.
- [x] Revisar diff final de archivos existentes modificados.
- [x] Registrar evidencia, riesgos residuales y siguiente owner en Result Contract.
