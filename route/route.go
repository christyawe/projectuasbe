package route

import (
	"database/sql"
	"uas_backend/helper"
	"uas_backend/middleware"
	model "uas_backend/model/Postgresql"
	"uas_backend/repository"
	"uas_backend/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(app *fiber.App, db *sql.DB, mongoColl *mongo.Collection) {
	api := app.Group("/api")

	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "API Running 🚀",
		})
	})

	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo)

	auth := api.Group("/auth")

	auth.Post("/login", func(c *fiber.Ctx) error {
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.Error(c, fiber.StatusBadRequest, "Invalid JSON body")
		}

		if req.Username == "" || req.Password == "" {
			return helper.Error(c, fiber.StatusBadRequest, "Username and password are required")
		}

		resp, err := authService.Login(&req)
		if err != nil {
			switch err.Error() {
			case "invalid username or password":
				return helper.Error(c, fiber.StatusUnauthorized, err.Error())
			case "account is inactive, please contact admin":
				return helper.Error(c, fiber.StatusForbidden, err.Error())
			default:
				return helper.Error(c, fiber.StatusInternalServerError, "Internal Server Error")
			}
		}

		return helper.Success(c, resp)
	})

	// Achievement routes
	achievementRepo := repository.NewAchievementRepository(db, mongoColl)
	achievementService := service.NewAchievementService(achievementRepo)

	achievements := api.Group("/achievements")
	achievements.Use(middleware.RBAC("")) // Require authentication

	// FR-004: Submit untuk Verifikasi
	achievements.Post("/:id/submit", func(c *fiber.Ctx) error {
		// Get achievement ID from URL params
		achievementIDStr := c.Params("id")
		achievementID, err := uuid.Parse(achievementIDStr)
		if err != nil {
			return helper.Error(c, fiber.StatusBadRequest, "Invalid achievement ID format")
		}

		// Get user info from JWT token (set by middleware)
		claims := c.Locals("user_info")
		if claims == nil {
			return helper.Error(c, fiber.StatusUnauthorized, "User info not found")
		}

		userIDStr, ok := claims.(map[string]interface{})["user_id"].(string)
		if !ok {
			return helper.Error(c, fiber.StatusUnauthorized, "Invalid user ID in token")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return helper.Error(c, fiber.StatusUnauthorized, "Invalid user ID format")
		}

		// Call service to submit for verification
		result, err := achievementService.SubmitForVerification(c.Context(), userID, achievementID)
		if err != nil {
			switch err.Error() {
			case "student data not found for this user":
				return helper.Error(c, fiber.StatusNotFound, err.Error())
			case "achievement not found":
				return helper.Error(c, fiber.StatusNotFound, err.Error())
			case "unauthorized: achievement does not belong to this student":
				return helper.Error(c, fiber.StatusForbidden, err.Error())
			case "achievement must be in 'draft' status to submit":
				return helper.Error(c, fiber.StatusBadRequest, err.Error())
			default:
				return helper.Error(c, fiber.StatusInternalServerError, "Failed to submit achievement")
			}
		}

		return helper.Success(c, fiber.Map{
			"message":     "Achievement submitted for verification successfully",
			"achievement": result,
		})
	})

	// TODO: User routes
}
