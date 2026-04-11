package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type schema struct {
	Tables        map[string]*table
	Relationships []relationship
}

type table struct {
	Name    string
	Columns []column
}

type column struct {
	Name string
	Type string
	PK   bool
	FK   bool
}

type relationship struct {
	FromTable  string
	FromColumn string
	ToTable    string
	ToColumn   string
}

func main() {
	input := flag.String("input", "../../infra/migrations/000001_init_schema.up.sql", "path to PostgreSQL schema migration")
	flag.Parse()

	raw, err := os.ReadFile(*input)
	if err != nil {
		exitErr("read schema: %v", err)
	}

	parsed := parseSchema(string(raw))
	output := renderMermaid(parsed)
	fmt.Print(output)
}

func exitErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func parseSchema(raw string) schema {
	clean := stripLineComments(raw)
	parsed := schema{Tables: map[string]*table{}}

	for _, block := range createTableBlocks(clean) {
		tbl := table{Name: block.name}
		for _, item := range splitTopLevel(block.body, ',') {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			upper := strings.ToUpper(item)
			switch {
			case strings.HasPrefix(upper, "CONSTRAINT") && strings.Contains(upper, "FOREIGN KEY"):
				if rel, ok := parseTableForeignKey(block.name, item); ok {
					parsed.Relationships = append(parsed.Relationships, rel)
				}
				continue
			case strings.HasPrefix(upper, "UNIQUE"),
				strings.HasPrefix(upper, "PRIMARY KEY"),
				strings.HasPrefix(upper, "CHECK"),
				strings.HasPrefix(upper, "EXCLUDE"):
				continue
			}

			col, rel, hasRel, ok := parseColumn(block.name, item)
			if !ok {
				continue
			}
			tbl.Columns = append(tbl.Columns, col)
			if hasRel {
				parsed.Relationships = append(parsed.Relationships, rel)
			}
		}
		parsed.Tables[tbl.Name] = &tbl
	}

	parsed.Relationships = append(parsed.Relationships, parseAlterForeignKeys(clean)...)
	markForeignKeyColumns(&parsed)
	sortSchema(&parsed)
	return parsed
}

type createTableBlock struct {
	name string
	body string
}

func createTableBlocks(sql string) []createTableBlock {
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	matches := re.FindAllStringSubmatchIndex(sql, -1)
	blocks := make([]createTableBlock, 0, len(matches))
	for _, match := range matches {
		name := sql[match[2]:match[3]]
		openParen := match[1] - 1
		closeParen := findMatchingParen(sql, openParen)
		if closeParen == -1 {
			continue
		}
		blocks = append(blocks, createTableBlock{
			name: name,
			body: sql[openParen+1 : closeParen],
		})
	}
	return blocks
}

func findMatchingParen(s string, open int) int {
	depth := 0
	inSingleQuote := false
	for i := open; i < len(s); i++ {
		ch := s[i]
		if ch == '\'' {
			inSingleQuote = !inSingleQuote
			continue
		}
		if inSingleQuote {
			continue
		}
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func splitTopLevel(s string, sep rune) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	inSingleQuote := false
	for _, r := range s {
		if r == '\'' {
			inSingleQuote = !inSingleQuote
			current.WriteRune(r)
			continue
		}
		if !inSingleQuote {
			switch r {
			case '(':
				depth++
			case ')':
				depth--
			default:
				if r == sep && depth == 0 {
					parts = append(parts, current.String())
					current.Reset()
					continue
				}
			}
		}
		current.WriteRune(r)
	}
	parts = append(parts, current.String())
	return parts
}

func parseColumn(tableName string, item string) (column, relationship, bool, bool) {
	fields := strings.Fields(item)
	if len(fields) < 2 {
		return column{}, relationship{}, false, false
	}

	name := trimIdentifier(fields[0])
	if name == "" {
		return column{}, relationship{}, false, false
	}

	rest := strings.TrimSpace(item[len(fields[0]):])
	typeName := columnType(rest)
	if typeName == "" {
		typeName = "unknown"
	}

	col := column{
		Name: name,
		Type: normalizeType(typeName),
		PK:   containsWord(rest, "PRIMARY") && containsWord(rest, "KEY"),
	}

	rel, hasRel := inlineForeignKey(tableName, name, rest)
	col.FK = hasRel
	return col, rel, hasRel, true
}

func columnType(rest string) string {
	fields := strings.Fields(rest)
	var parts []string
	for _, field := range fields {
		upper := strings.ToUpper(strings.Trim(field, ","))
		switch upper {
		case "PRIMARY", "UNIQUE", "NOT", "NULL", "DEFAULT", "REFERENCES", "CHECK", "GENERATED", "COLLATE", "CONSTRAINT":
			return strings.Join(parts, " ")
		}
		parts = append(parts, field)
	}
	return strings.Join(parts, " ")
}

func inlineForeignKey(tableName string, columnName string, rest string) (relationship, bool) {
	re := regexp.MustCompile(`(?i)\bREFERENCES\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)`)
	match := re.FindStringSubmatch(rest)
	if len(match) == 0 {
		return relationship{}, false
	}
	return relationship{
		FromTable:  tableName,
		FromColumn: columnName,
		ToTable:    match[1],
		ToColumn:   match[2],
	}, true
}

func parseTableForeignKey(tableName string, item string) (relationship, bool) {
	re := regexp.MustCompile(`(?i)FOREIGN\s+KEY\s*\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)\s+REFERENCES\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)`)
	match := re.FindStringSubmatch(item)
	if len(match) == 0 {
		return relationship{}, false
	}
	return relationship{
		FromTable:  tableName,
		FromColumn: match[1],
		ToTable:    match[2],
		ToColumn:   match[3],
	}, true
}

func parseAlterForeignKeys(sql string) []relationship {
	re := regexp.MustCompile(`(?is)ALTER\s+TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+ADD\s+CONSTRAINT\s+[a-zA-Z_][a-zA-Z0-9_]*\s+FOREIGN\s+KEY\s*\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)\s+REFERENCES\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)`)
	matches := re.FindAllStringSubmatch(sql, -1)
	rels := make([]relationship, 0, len(matches))
	for _, match := range matches {
		rels = append(rels, relationship{
			FromTable:  match[1],
			FromColumn: match[2],
			ToTable:    match[3],
			ToColumn:   match[4],
		})
	}
	return rels
}

func renderMermaid(parsed schema) string {
	var b strings.Builder
	b.WriteString("erDiagram\n")
	for _, rel := range parsed.Relationships {
		b.WriteString(fmt.Sprintf("  %s ||--o{ %s : %s_to_%s\n",
			mermaidName(rel.ToTable),
			mermaidName(rel.FromTable),
			mermaidLabel(rel.FromColumn),
			mermaidLabel(rel.ToColumn),
		))
	}

	for _, name := range tableNames(parsed.Tables) {
		tbl := parsed.Tables[name]
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s {\n", mermaidName(tbl.Name)))
		for _, col := range tbl.Columns {
			markers := []string{}
			if col.PK {
				markers = append(markers, "PK")
			}
			if col.FK {
				markers = append(markers, "FK")
			}
			marker := ""
			if len(markers) > 0 {
				marker = " " + strings.Join(markers, ",")
			}
			b.WriteString(fmt.Sprintf("    %s %s%s\n", mermaidType(col.Type), mermaidName(col.Name), marker))
		}
		b.WriteString("  }\n")
	}

	return b.String()
}

func tableNames(tables map[string]*table) []string {
	names := make([]string, 0, len(tables))
	for name := range tables {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortSchema(parsed *schema) {
	for _, tbl := range parsed.Tables {
		sort.SliceStable(tbl.Columns, func(i, j int) bool {
			return tbl.Columns[i].Name < tbl.Columns[j].Name
		})
	}

	seen := map[string]struct{}{}
	rels := make([]relationship, 0, len(parsed.Relationships))
	for _, rel := range parsed.Relationships {
		key := strings.Join([]string{rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn}, ".")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		rels = append(rels, rel)
	}
	sort.SliceStable(rels, func(i, j int) bool {
		left := strings.Join([]string{rels[i].ToTable, rels[i].FromTable, rels[i].FromColumn}, ".")
		right := strings.Join([]string{rels[j].ToTable, rels[j].FromTable, rels[j].FromColumn}, ".")
		return left < right
	})
	parsed.Relationships = rels
}

func markForeignKeyColumns(parsed *schema) {
	for _, rel := range parsed.Relationships {
		tbl, ok := parsed.Tables[rel.FromTable]
		if !ok {
			continue
		}
		for i := range tbl.Columns {
			if tbl.Columns[i].Name == rel.FromColumn {
				tbl.Columns[i].FK = true
			}
		}
	}
}

func stripLineComments(raw string) string {
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		lines[i] = stripLineComment(line)
	}
	return strings.Join(lines, "\n")
}

func stripLineComment(line string) string {
	inSingleQuote := false
	for i := 0; i < len(line)-1; i++ {
		if line[i] == '\'' {
			inSingleQuote = !inSingleQuote
			continue
		}
		if !inSingleQuote && line[i] == '-' && line[i+1] == '-' {
			return line[:i]
		}
	}
	return line
}

func trimIdentifier(raw string) string {
	return strings.Trim(raw, `" ,`)
}

func normalizeType(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, ",")
	raw = regexp.MustCompile(`\s+`).ReplaceAllString(raw, " ")
	return raw
}

func mermaidType(raw string) string {
	raw = strings.ToLower(raw)
	raw = strings.ReplaceAll(raw, " ", "_")
	raw = strings.ReplaceAll(raw, "[]", "_array")
	raw = regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(raw, "")
	raw = strings.Trim(raw, "_")
	if raw == "" {
		return "unknown"
	}
	return raw
}

func mermaidName(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(unicode.ToUpper(r))
			continue
		}
		b.WriteRune('_')
	}
	return b.String()
}

func mermaidLabel(raw string) string {
	return strings.ToLower(mermaidName(raw))
}

func containsWord(raw string, word string) bool {
	re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
	return re.MatchString(raw)
}
