package test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"UASBE/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	app := fiber.New()

	t.Run("Success with data", func(t *testing.T) {
		app.Get("/test", func(c *fiber.Ctx) error {
			return helper.Success(c, fiber.Map{
				"message": "test success",
				"id":      123,
			})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		assert.Equal(t, "success", result["status"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Success with nil data", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/test", func(c *fiber.Ctx) error {
			return helper.Success(c, nil)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app2.Test(req)

		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		assert.Equal(t, "success", result["status"])
	})

	t.Run("Success with array data", func(t *testing.T) {
		app3 := fiber.New()
		app3.Get("/test", func(c *fiber.Ctx) error {
			return helper.Success(c, []string{"item1", "item2", "item3"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app3.Test(req)

		assert.Equal(t, 200, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		assert.Equal(t, "success", result["status"])
		assert.NotNil(t, result["data"])
	})
}

func TestError(t *testing.T) {
	app := fiber.New()

	t.Run("Error 400", func(t *testing.T) {
		app.Get("/test", func(c *fiber.Ctx) error {
			return helper.Error(c, 400, "Bad request")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		assert.Equal(t, "error", result["status"])
		assert.Equal(t, "Bad request", result["message"])
	})

	t.Run("Error 401", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/test", func(c *fiber.Ctx) error {
			return helper.Error(c, 401, "Unauthorized")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app2.Test(req)

		assert.Equal(t, 401, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		assert.Equal(t, "error", result["status"])
		assert.Equal(t, "Unauthorized", result["message"])
	})

	t.Run("Error 500", func(t *testing.T) {
		app3 := fiber.New()
		app3.Get("/test", func(c *fiber.Ctx) error {
			return helper.Error(c, 500, "Internal server error")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app3.Test(req)

		assert.Equal(t, 500, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		assert.Equal(t, "error", result["status"])
		assert.Equal(t, "Internal server error", result["message"])
	})

	t.Run("Error with empty message", func(t *testing.T) {
		app4 := fiber.New()
		app4.Get("/test", func(c *fiber.Ctx) error {
			return helper.Error(c, 404, "")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app4.Test(req)

		assert.Equal(t, 404, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		assert.Equal(t, "error", result["status"])
		assert.Equal(t, "", result["message"])
	})
}
