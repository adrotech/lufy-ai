# Manual visual de Lufy

Guía HTML del funcionamiento del harness inline: cómo pedir trabajo, cómo avanza el principal, cuándo consulta, cuándo puede delegar, qué aportan las skills y cómo se reconoce el resultado.

![Portada del manual visual de Lufy](preview.png)

## Abrirlo

Abrí [index.html](index.html) en un navegador desde un checkout del repositorio. Conservá `manual.css`, `scenarios.js`, `manual.js` y el banner en `docs/assets/brand/`: los recursos usan rutas relativas. No necesita build, paquetes, fuentes externas ni red para funcionar localmente.

Si preferís una vista HTTP, desde la raíz del repositorio:

```sh
python3 -m http.server 8766 --bind 127.0.0.1
```

Abrí `http://127.0.0.1:8766/docs/manual/`. El puerto es un ejemplo; elegí otro si ya está ocupado. Detené el servidor con Ctrl+C al terminar.

GitHub muestra el código del HTML. Versionar esta guía no habilita GitHub Pages ni un hosting público. Para verla descargada, necesitás conservar los archivos y assets relativos del checkout.

## Qué permite explorar

- Ocho recorridos: comprender código, corregir, agregar una funcionalidad, crear desde cero, proponer, redactar/traducir, analizar datos y operar/recuperar.
- Pedidos editables y copiables, con indicación de resultado, contexto y restricciones.
- Cada paso muestra conversación ilustrativa, acción, evidencia pertinente y capacidades que podrían ayudar.
- Respuestas finales de ejemplo, condición de término y límites del alcance.
- Principal, usuario, subagentes condicionales y cinco skills reales del catálogo.
- Un visor de delegación que combina `off`, `on_request`, `auto`, herramientas del host e instrucciones del usuario.
- Navegación con teclado, controles accesibles, adaptación a pantallas pequeñas y preferencia de movimiento reducido.

Los personajes se representan con recortes visuales CSS del banner original: no se agregan imágenes ni se atribuyen roles instalados a los avatares.

## Fuentes y límites

La guía explica las políticas actuales de [trabajo](../../.lufy/contracts/work.md), [evidencia](../../.lufy/contracts/evidence.md), [entrega](../../.lufy/contracts/delivery.md) y [configuración](../configuration.md). Los ejemplos son didácticos con datos ficticios; no son trazas de modelos, pruebas reales ni una medición de calidad. Ningún control ejecuta el harness o modifica su configuración.

Una skill es una guía; no es otro agente. La delegación depende del host, la política efectiva y el alcance del pedido. El principal conserva la integración y las comprobaciones. El hook de sesión orienta con metadata local; no registra automáticamente conversaciones, runs o reviews.

## Mantenerlo

- `index.html`: estructura y explicación estática.
- `manual.css`: presentación, retratos del banner, responsive y estilos de impresión.
- `scenarios.js`: recorridos, conversaciones y resultados ilustrativos.
- `manual.js`: navegación de casos/pasos, copia del pedido y visor de delegación.

Al cambiar una política, revisar el contenido y los ejemplos afectados. Comprobar enlaces, sintaxis JavaScript, recorrido de los ocho casos, estado de los controles, reglas de delegación y vistas de escritorio/móvil. Usar el navegador disponible para la verificación visual; los checks estáticos no prueban la interacción.

No es un asset instalado por `sync` y no modifica el catálogo embedded. Para volver a las guías de operación, consultá [el índice de documentación](../README.md).

## Verificación de esta revisión

La revisión del 29/10/2026 se contrastó con las políticas del flujo inline y se comprobó en navegador mediante HTTP local:

- Ocho casos y 36 pasos: navegación directa, anterior/siguiente, límites y resultados finales.
- Ocho estados del visor de delegación, incluyendo falta de herramientas y prioridad de la instrucción inline.
- Teclado en las pestañas, cambio de ejemplo del editor, pedido vacío y copia al portapapeles.
- Ejemplo de intercambio con tareas delegadas y respuesta parcial conservada como tal.
- Vista a 390 px de ancho dentro de un viewport embebido: avance, lectura y foco del nuevo paso. No representa un dispositivo físico.
- Sin errores o warnings en los logs observados del navegador; respuesta HTTP 200.
- Sintaxis de ambos archivos JavaScript, 54 IDs HTML únicos, 32 referencias HTML y recursos locales comprobados. Sin assets externos.

La apertura directa mediante `file://` no pudo ensayarse en el navegador integrado: su política sólo admite HTTP/HTTPS. La vista HTTP y los recursos locales sí se comprobaron. Los screenshots y salidas de QA se conservaron como evidencia temporal fuera del repositorio, en `/private/tmp/lufy-manual-20261029`.

La validación agrupada `scripts/validate.sh` pasó con 63 pruebas Python offline, cobertura Go 80,6%, vet, build, paridad, workflows y PR guard. ShellCheck no está instalado localmente; CI aplica ese gate. Estas comprobaciones del repositorio se distinguen de la verificación funcional del HTML descrita arriba.
