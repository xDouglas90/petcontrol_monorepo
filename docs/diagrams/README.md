# Diagrams

This directory stores architecture and data-model diagrams for PetControl.

## Entity-Relationship Diagram

- Source schema: `infra/migrations/000001_init_schema.up.sql`
- Generated diagram: `docs/diagrams/er-diagram.mmd`
- Generator: `apps/api/cmd/erdiagram`
- Validator: `go-mermaid`

The current Mermaid file is generated from the initial schema migration and
includes all tables and foreign-key relationships detected in that file.

Regenerate after schema changes:

```bash
make diagrams
```

Validate the Mermaid syntax with `go-mermaid`:

```bash
make diagrams-check
```
