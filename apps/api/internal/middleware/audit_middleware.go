package middleware

import (
	"encoding/json"
	"log/slog"
	"net/netip"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

const auditEntriesContextKey = "audit_entries"

type AuditEntry struct {
	Action      sqlc.LogAction
	EntityTable string
	EntityID    pgtype.UUID
	CompanyID   pgtype.UUID
	OldData     any
	NewData     any
}

func AddAuditEntry(c *gin.Context, entry AuditEntry) {
	entries, _ := c.Get(auditEntriesContextKey)
	current, _ := entries.([]AuditEntry)
	current = append(current, entry)
	c.Set(auditEntriesContextKey, current)
}

func Audit(queries sqlc.Querier, logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= 400 {
			return
		}

		raw, ok := c.Get(auditEntriesContextKey)
		if !ok {
			return
		}

		entries, ok := raw.([]AuditEntry)
		if !ok || len(entries) == 0 {
			return
		}

		var changedBy pgtype.UUID
		if claims, claimsOK := GetClaims(c); claimsOK {
			if parsed, err := parseUUID(claims.UserID); err == nil {
				changedBy = parsed
			}
		}

		var ipAddress netip.Addr
		if parsedIP, err := netip.ParseAddr(c.ClientIP()); err == nil {
			ipAddress = parsedIP
		}

		for _, entry := range entries {
			newData := mustJSON(entry.NewData)
			if len(newData) == 0 {
				newData = []byte("{}")
			}

			params := sqlc.InsertAuditLogParams{
				Action:      entry.Action,
				EntityTable: entry.EntityTable,
				EntityID:    entry.EntityID,
				CompanyID:   entry.CompanyID,
				OldData: func() []byte {
					if entry.OldData == nil {
						return nil
					}
					return mustJSON(entry.OldData)
				}(),
				NewData:   newData,
				ChangedBy: changedBy,
				IPAddress: func() *netip.Addr {
					if !ipAddress.IsValid() {
						return nil
					}
					value := ipAddress
					return &value
				}(),
				UserAgent: pgtype.Text{String: c.Request.UserAgent(), Valid: c.Request.UserAgent() != ""},
			}

			if err := queries.InsertAuditLog(c.Request.Context(), params); err != nil {
				logger.Error("failed to persist audit log",
					"correlation_id", GetCorrelationID(c),
					"entity_table", entry.EntityTable,
					"entity_id", entry.EntityID.String(),
					"error", err.Error(),
				)
			}
		}
	}
}

func mustJSON(value any) []byte {
	if value == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return []byte("{}")
	}
	return payload
}
