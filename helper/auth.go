package helper

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetUserIDFromContext mengambil user_id dari JWT claims yang disimpan di context
func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	claims := c.Locals("user_info")
	if claims == nil {
		return uuid.Nil, errors.New("user info not found")
	}

	userIDStr, ok := claims.(map[string]interface{})["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID format")
	}

	return userID, nil
}
