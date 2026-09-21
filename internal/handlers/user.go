package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"magtapp-backend/internal/dto"
	"magtapp-backend/internal/middleware"
)

// Me returns the caller's user record. The "create or identify" logic
// itself already ran in the Auth middleware before this handler is ever
// reached — this endpoint just reflects the result back.
func Me(c echo.Context) error {
	u := middleware.FromContext(c)
	return c.JSON(http.StatusOK, dto.MeResponse{
		ID:        u.ID.String(),
		CreatedAt: u.CreatedAt,
	})
}
