package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
)

func TestAuditMiddleware_PersistsAuditEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	companyID := newPGUUID(t)
	entityID := newPGUUID(t)
	userID := uuid.NewString()

	mock.ExpectExec(`(?s)name: InsertAuditLog`).
		WithArgs(
			sqlc.LogActionCreate,
			"schedules",
			entityID,
			companyID,
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			mustParseUUID(t, userID),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	router := gin.New()
	router.Use(RequestContext(), Audit(queries, nil))
	router.POST("/mutate", func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{
			UserID:    userID,
			CompanyID: companyID.String(),
			Role:      "admin",
			Kind:      "owner",
		})

		AddAuditEntry(c, AuditEntry{
			Action:      sqlc.LogActionCreate,
			EntityTable: "schedules",
			EntityID:    entityID,
			CompanyID:   companyID,
			OldData:     nil,
			NewData:     gin.H{"id": entityID.String(), "status": "waiting", "ts": time.Now().UTC()},
		})

		JSONData(c, http.StatusCreated, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/mutate", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
