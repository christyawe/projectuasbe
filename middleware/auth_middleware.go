package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"uas_backend/helper"
	"uas_backend/utils"
)

func RBAC(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.Error(c, fiber.StatusUnauthorized, "Missing Authorization Header")
		}

		tokenString := utils.ExtractToken(authHeader)
		if tokenString == "" {
			return helper.Error(c, fiber.StatusUnauthorized, "Invalid Token Format")
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			return helper.Error(c, fiber.StatusUnauthorized, "Invalid or Expired Token")
		}

		c.Locals("user_info", claims)

		if requiredPermission == "" {
			return c.Next()
		}

		userPermissionsInterface, ok := claims["permissions"].([]interface{})
		if !ok {
			return helper.Error(c, fiber.StatusForbidden, "Invalid permissions data in token")
		}

		hasPermission := false
		for _, p := range userPermissionsInterface {
			if pStr, ok := p.(string); ok {
				if strings.EqualFold(pStr, requiredPermission) {
					hasPermission = true
					break
				}
			}
		}

		if !hasPermission {
			return helper.Error(c, fiber.StatusForbidden, "You do not have permission: "+requiredPermission)
		}

		return c.Next()
	}
}