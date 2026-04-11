package main

import (
	"strings"
	"testing"
)

func TestParseSchemaIncludesInlineAndAlterForeignKeys(t *testing.T) {
	sql := `
CREATE TABLE users (
  id UUID PRIMARY KEY
);

CREATE TABLE companies (
  id UUID PRIMARY KEY,
  responsible_id UUID NOT NULL
);

CREATE TABLE company_users (
  id UUID PRIMARY KEY,
  company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE (company_id, user_id)
);

ALTER TABLE companies
  ADD CONSTRAINT fk_companies_responsible
  FOREIGN KEY (responsible_id) REFERENCES users(id) ON DELETE RESTRICT;
`

	parsed := parseSchema(sql)

	if got := len(parsed.Tables); got != 3 {
		t.Fatalf("expected 3 tables, got %d", got)
	}
	if got := len(parsed.Relationships); got != 3 {
		t.Fatalf("expected 3 relationships, got %d: %#v", got, parsed.Relationships)
	}

	rendered := renderMermaid(parsed)
	for _, expected := range []string{
		"COMPANIES ||--o{ COMPANY_USERS : company_id_to_id",
		"USERS ||--o{ COMPANY_USERS : user_id_to_id",
		"USERS ||--o{ COMPANIES : responsible_id_to_id",
		"uuid COMPANY_ID FK",
		"uuid RESPONSIBLE_ID FK",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("expected rendered diagram to contain %q\n%s", expected, rendered)
		}
	}
}
