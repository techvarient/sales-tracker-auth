package middleware

import (
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/sales-tracker/auth-service/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/sales-tracker/auth-service/internal/domain"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			
			start := time.Now()
			defer func() {
				latency := time.Since(start)
				logrus.Infof("%s %s %d %v", req.Method, req.URL.Path, res.Status, latency)
			}()
			
			return next(c)
		}
	}
}

func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			d := recover()
			if d != nil {
				logrus.Error("Recovered from panic:", d)
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
			return next(c)
		}
	}
}

func CORS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusNoContent)
			}
			
			return next(c)
		}
	}
}

func JWTMiddleware(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authorization := c.Request().Header.Get("Authorization")
			if authorization == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "No authorization header provided")
			}

			tokenString := strings.Replace(authorization, "Bearer ", "", 1)
			claims := &domain.JWTClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.JWTSecret), nil
			})

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			if token.Valid {
				c.Set("user_id", claims.UserID)
				c.Set("role", claims.Role)
				c.Set("email", claims.Email)
				return next(c)
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
		}
	}
}

func RoleMiddleware(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get("role").(string)
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}
}
