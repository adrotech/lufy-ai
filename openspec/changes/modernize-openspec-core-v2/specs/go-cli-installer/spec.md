## ADDED Requirements

### Requirement: CLI instala workflow OpenSpec core v2 standalone
La CLI Go SHALL instalar el workflow OpenSpec core v2 desde checkout de desarrollo o desde assets embebidos de release sin requerir clone adicional.

#### Scenario: Instalación desde release incluye core v2
- **WHEN** el usuario ejecuta `lufy-ai install --target <dir>` con un binario release que contiene assets core v2 embebidos
- **THEN** el target recibe la configuración, comandos, skills y baseline OpenSpec core v2 gestionados

#### Scenario: Sync desde release actualiza core v2
- **WHEN** el usuario ejecuta `lufy-ai sync --target <dir>` con un binario release que contiene assets core v2 embebidos
- **THEN** la CLI compara el catálogo embebido y planifica updates seguros para assets OpenSpec core v2 gestionados

### Requirement: CLI mantiene validación de paridad de assets OpenSpec
La implementación SHALL mantener validación automática para evitar que specs o comandos OpenSpec raíz diverjan de sus copias embebidas.

#### Scenario: Tests detectan drift de assets embebidos
- **WHEN** un asset OpenSpec core v2 raíz cambia sin actualizar su copia embebida
- **THEN** los tests Go de assets fallan indicando drift entre catálogo raíz y embebido

#### Scenario: Validación agrupada ejecuta paridad relevante
- **WHEN** se ejecuta `scripts/validate.sh`
- **THEN** la validación Go incluye la comprobación de paridad entre assets raíz y embebidos
