package test

import (
	"testing"
	"UASBE/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUserIDFromContext(t *testing.T) {
	app := fiber.New()

	t.Run("Valid user_id in context", func(t *testing.T) {
		app.Get("/test", func(c *fiber.Ctx) error {
			userID := uuid.New()
			c.Locals("user_info", map[string]interface{}{
				"user_id": userID.String(),
			})

			result, err := helper.GetUserIDFromContext(c)
			assert.NoError(t, err)
			assert.Equal(t, userID, result)
			return nil
		})
	})

	t.Run("No user_info in context", func(t *testing.T) {
		app.Get("/test2", func(c *fiber.Ctx) error {
			result, err := helper.GetUserIDFromContext(c)
			assert.Error(t, err)
			assert.Equal(t, "user info not found", err.Error())
			assert.Equal(t, uuid.Nil, result)
			return nil
		})
	})

	t.Run("Invalid user_id format", func(t *testing.T) {
		app.Get("/test3", func(c *fiber.Ctx) error {
			c.Locals("user_info", map[string]interface{}{
				"user_id": "invalid-uuid",
			})

			result, err := helper.GetUserIDFromContext(c)
			assert.Error(t, err)
			assert.Equal(t, "invalid user ID format", err.Error())
			assert.Equal(t, uuid.Nil, result)
			return nil
		})
	})

	t.Run("Missing user_id in claims", func(t *testing.T) {
		app.Get("/test4", func(c *fiber.Ctx) error {
			c.Locals("user_info", map[string]interface{}{
				"username": "testuser",
			})

			result, err := helper.GetUserIDFromContext(c)
			assert.Error(t, err)
			assert.Equal(t, "invalid user ID in token", err.Error())
			assert.Equal(t, uuid.Nil, result)
			return nil
		})
	})
}

func TestExtractTokenFromHeader(t *testing.T) {
	app := fiber.New()

	t.Run("Valid Bearer token", func(t *testing.T) {
		app.Get("/test", func(c *fiber.Ctx) error {
			c.Request().Header.Set("Authorization", "Bearer test.token.here")

			token, err := helper.ExtractTokenFromHeader(c)
			assert.NoError(t, err)
			assert.Equal(t, "test.token.here", token)
			return nil
		})
	})

	t.Run("No Authorization header", func(t *testing.T) {
		app.Get("/test2", func(c *fiber.Ctx) error {
			token, err := helper.ExtractTokenFromHeader(c)
			assert.Error(t, err)
			assert.Equal(t, "authorization header not found", err.Error())
			assert.Empty(t, token)
			return nil
		})
	})

	t.Run("Invalid header format - no Bearer prefix", func(t *testing.T) {
		app.Get("/test3", func(c *fiber.Ctx) error {
			c.Request().Header.Set("Authorization", "test.token.here")

			token, err := helper.ExtractTokenFromHeader(c)
			assert.Error(t, err)
			assert.Equal(t, "invalid authorization header format", err.Error())
			assert.Empty(t, token)
			return nil
		})
	})

	t.Run("Empty token after Bearer", func(t *testing.T) {
		app.Get("/test4", func(c *fiber.Ctx) error {
			c.Request().Header.Set("Authorization", "Bearer ")

			token, err := helper.ExtractTokenFromHeader(c)
			assert.Error(t, err)
			assert.Equal(t, "token is empty", err.Error())
			assert.Empty(t, token)
			return nil
		})
	})

	t.Run("Short Authorization header", func(t *testing.T) {
		app.Get("/test5", func(c *fiber.Ctx) error {
			c.Request().Header.Set("Authorization", "Bear")

			token, err := helper.ExtractTokenFromHeader(c)
			assert.Error(t, err)
			assert.Equal(t, "invalid authorization header format", err.Error())
			assert.Empty(t, token)
			return nil
		})
	})
}
