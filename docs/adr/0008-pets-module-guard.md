# ADR 0008: Guarda do domínio Pets pelo módulo CRM

## Status

Aceito

## Contexto

Durante a expansão do núcleo operacional do PetControl, foi implementado o CRUD para as entidades `clients`, `pets` e `services`.
Observou-se uma discrepância semântica em relação a alguns registros documentais anteriores e lógicos:

- O código atual e partes documentais planeadas protegem as rotas de `pets` requerendo o módulo `CRM` ativo (`RequireModule(queries, "CRM")`).
- Semanticamente, `pets` também estão fortemente ligados ao contexto de atendimento e agendamentos (`SCH`), o que originou questionamentos de domínio.

O conflito central reside na pergunta: o modelo de negócio vê "Pets" primariamente como cadastro relacional (clientes e agregados) ou como itens operacionais de atendimento?

## Decisão

O domínio `pets` continuará sob a guarda primária do módulo `CRM` (Customer Relationship Management) tanto na infraestrutura de roteamento do Backend quanto na conceituação arquitetural.

### Justificativas

1. **Modelagem Baseada em Relacionamento**: Um pet (paciente) existe unicamente como uma entidade subordinada a um `client` (tutor), com chave estrangeira obrigatória (`owner_id` referenciando `clients.id`).
2. **Separação de Preocupações (SoC)**: O módulo `SCH` (Agendamentos) foca em transações de tempo e serviços. O gerenciamento de portfólio de entes (tutores e pets) e atributos duradouros deles (aniversários, raça, etc) pertence categoricamente ao ciclo de relacionamento (`CRM`).
3. **Agrupamento de Venda Comercial**: Os pacotes vigentes assumem uma estrutura fundamental de licença combinada onde a contratação de pacotes abrange módulos base em tandem como "Core Operacional".

## Consequências

- **Acesso Obrigatório em Cascatas**: Para que o módulo `SCH` funcione e associe cliente/pet a agendamentos, o tenant deve preferencialmente possuir acesso ao módulo `CRM` para realizar o registro e a consulta prévia de tais entidades.
- **Acoplamento Inerente de Domínio**: Adota-se que o módulo `SCH` possui uma forte dependência *business-logic-wise* sobre domínios providos pelo `CRM`.
- Futuramente, se a API passar a vender licenças extremante fracionadas e for demandado que a gestão de pets ocorra unicamente pelo aspecto clínico sem o viés de retenção de clientes do `CRM`, essa guarda modular nas rotas deverá ser extraída para verificações polimórficas (ex: `RequireAnyModule("CRM", "CLINIC")`). No cenário em vigência, o acoplamento de autorização é deliberado.
