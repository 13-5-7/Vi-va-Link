package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTAuth(jwtSecret string) echo.MiddlewareFunc {
	log.Println("-----JWTAuth called-----")
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "missing or invalid authorization header",
					},
				})
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "invalid or expired token",
					},
				})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "invalid token claims",
					},
				})
			}

			c.Set("user_id", fmt.Sprintf("%v", claims["user_id"]))
			c.Set("role", fmt.Sprintf("%v", claims["role"]))
			return next(c)
		}
	}
}

func RequireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole, ok := c.Get("role").(string)
			if !ok || userRole != role {
				return c.JSON(http.StatusForbidden, map[string]any{
					"error": map[string]string{
						"code":    "FORBIDDEN",
						"message": "insufficient permissions",
					},
				})
			}
			return next(c)
		}
	}
}
