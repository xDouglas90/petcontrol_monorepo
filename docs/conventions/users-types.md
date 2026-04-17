# Tipos de usuários

## Índice

- [Tipos de usuários](#tipos-de-usuários)
  - [Índice](#índice)
  - [Tipos de usuário no sistema (papel técnico/sistêmico)](#tipos-de-usuário-no-sistema-papel-técnicosistêmico)
    - [root](#root)
    - [internal](#internal)
    - [admin](#admin)
    - [system](#system)
    - [common](#common)
    - [free](#free)
  - [Tipos de usuário no contexto de negócio](#tipos-de-usuário-no-contexto-de-negócio)
    - [owner](#owner)
    - [employee](#employee)
    - [client](#client)
    - [supplier](#supplier)
    - [outsourced\_employee](#outsourced_employee)

## Tipos de usuário no sistema (papel técnico/sistêmico)

*`user_role_type`*

### root

Usuário com acesso total ao sistema. Inserido no banco de dados apenas durante a execução do script de criação do banco de dados, ou, por outro usuário `root`.

### internal

Usuários para os desenvolvedores da aplicação. Criado por um usuário do tipo `root`.

Estes usuários não possuem uma empresa vinculada.

Poderão ter acesso a todos os tenants, mas não poderão criar novos tenants em produção, apenas em ambientes de desenvolvimento e testes.

### admin

Usuários para administradores de empresas(tenants). Criado por um usuário do tipo `root.

Poderá criar novos usuários, mas apenas para o tenant ao qual está vinculado.

Terá acesso total as configurações da empresa, gerenciamento de usuários e permissões a módulos do sistema.

Poderá trocar,cancelar e reativar planos de assinatura do tenant.

### system

Usuários criados por um `admin` para executar tarefas específicas do sistema.

Terão acesso somente aos módulos que foram liberados pelo `admin`.

### common

Usuários do tipo `common` são usuários comuns que podem ser criados por clientes dos tenants pela __interface de cadastro de novos usuários__, ou, por usuários do tipo `admin` e `system` com as devidas permissões.

Terão acesso somente aos módulos referentes a gerenciamento de pets(deles mesmos) e de agendamentos/serviços para os seus pets. Assim como de agenda disponível na semana.

### free

Usuários criados tanto por um `root` quanto pela __interface de cadastro de novos usuários__.

Não serão vinculados a nenhum tenant real, apenas ao futuro tenant `trial`, para demonstrações.

Estes usuários visualizarão o sistema por completo, com dados de demonstração(mocks). Poderão editar, criar e deletar dados, mas nada será salvo permanentemente.

## Tipos de usuário no contexto de negócio

*`user_kind`*

Estes usuários serão vinculados a um tenant na tabela `company_users`.

### owner

Proprietário/Sócio da empresa(tenant).

### employee

Funcionário da empresa cliente(tenant).

### client

Cliente da empresa cliente(tenant).

### supplier

Fornecedor da empresa cliente(tenant).

### outsourced_employee

Funcionário terceirizado(ex: free lancers, prestadores de serviço) da empresa cliente(tenant).
