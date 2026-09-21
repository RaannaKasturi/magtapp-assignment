package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"magtapp-backend/ent"
	"magtapp-backend/ent/user"
)

type contextKey string

const userContextKey contextKey = "authenticated_user"

// Auth implements the "basic Authorization: Bearer xyz" scheme the
// assignment asks for: the token itself IS the identity. If a user with
// this token exists, they're identified; if not, one is created on the
// spot. This is intentionally simple — in production this token would
// instead be a verified Firebase/JWT access token, and this middleware
// would validate its signature before ever touching the database.
func Auth(client *ent.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing or invalid Authorization header"})
			}
			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			if token == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "empty bearer token"})
			}

			ctx := c.Request().Context()
			u, err := client.User.Query().Where(user.TokenEQ(token)).Only(ctx)
			if ent.IsNotFound(err) {
				u, err = client.User.Create().SetToken(token).Save(ctx)
			}
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to resolve user"})
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(ctx, userContextKey, u)))
			return next(c)
		}
	}
}

// FromContext returns the user this request was authenticated as.
func FromContext(c echo.Context) *ent.User {
	u, _ := c.Request().Context().Value(userContextKey).(*ent.User)
	return u
}
