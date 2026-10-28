# Siguiente comparación: preparación, sin ejecuciones autorizadas aquí

El piloto de familias no se reutiliza como una comparación causal. El siguiente estudio necesita fijar el bundle efectivo de instrucciones de A y B, el modelo solicitado y el presupuesto antes de invocar candidatos. C queda identificado por los hashes actuales de política, casos y runner. Esto es una unidad propuesta; no un requisito de trabajo ordinario en Lufy.

- A: snapshot histórico del producto desde Git, sin mantener un motor legacy en producción. Debe representar su carga normal de contexto, no recibir documentos extra para sobrecargarlo. Si no puede ejecutarse con las restricciones comunes, se informa incompatibilidad, sin fingir una ejecución equivalente.
- B: principal inline con restricciones esenciales y alcance idéntico.
- C: principal inline con la política actual y capacidades pertinentes bajo demanda.
- Corpus exploratorio: seis fixtures existentes por variante, una ejecución por celda, sin retries para obtener verde; 18 entregas iniciales. Una muestra por celda sólo permite observación exploratoria.
- Orden: análisis A/B/C; propuesta B/C/A; escritura C/A/B; traducción A/B/C; corrección B/C/A; datos C/A/B. Workspaces nuevos y estado aislado por celda.
- Recursos propuestos: máximo 60 segundos por turno, sin paralelismo ni publicación. La infraestructura se comprueba sin modelos antes de iniciar. Ante fallo real se detiene el grupo; la recuperación se registra como nueva revisión de protocolo y conserva intentos/not_run anteriores.
- Segunda modificación: filtro —permitir una colección de estados conservando None y orden—; guía —cambiar el archivo de salida a gastos.csv y conservar el significado restante—. Partir de una copia de la entrega de cada variante y conservar el resultado anterior. Seis turnos adicionales, 24 en total como techo propuesto.
- Oráculos: validación funcional fuera del fixture y en sandbox; para texto, comparación factual/editorial contra la fuente, explícitamente revisada. La procedencia, permisos y archivos propios se comprueban en todas las celdas. No introducir assertions ocultas en el workspace del candidato.
- Registrar reloj monotónico, usage completo cuando exista, resultados, abandonos y costo humano de revisar/corregir/recuperar; conservar null cuando no se mida. Una reducción de bytes de prompt no es ahorro de tokens ni de costo.
- Aceptación previa: ninguna violación crítica observada y ninguna incógnita de controles críticos se considera resuelta. Investigar cualquier regresión de aceptación de C frente a B/A por familia; ahorro de recursos no la compensa. La segunda modificación debe preservar lo anterior. Si costo humano o condiciones no son comparables, no declarar adopción por comparación. No hay precisión estadística para superioridad general con esta muestra.

La revisión exacta de A/B, el modelo y el registro de tiempo/intervenciones aún requieren preparación antes de lanzar este estudio. La autorización recibida en esta sesión cubrió los cuatro casos del piloto, no estas 24 invocaciones adicionales.
