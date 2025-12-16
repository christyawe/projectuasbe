package service

import (
	"context"
	"errors"

	model "UASBE/model/Postgresql"
	"UASBE/repository"
	"UASBE/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService interface {
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	Logout(ctx context.Context, token string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.ProfileData, error)

	// HTTP endpoints
	LoginEndpoint(c *fiber.Ctx) error
	LogoutEndpoint(c *fiber.Ctx) error
	ProfileEndpoint(c *fiber.Ctx) error
}

type authService struct {
	authRepo *repository.AuthRepository
}

func NewAuthService(authRepo *repository.AuthRepository) AuthService {
	return &authService{authRepo: authRepo}
}

// Helper function untuk mengekstrak user ID dari JWT claims
func extractUserIDFromClaimsAuth(c *fiber.Ctx) (uuid.UUID, error) {
	claims := c.Locals("user_info")
	if claims == nil {
		return uuid.Nil, errors.New("user info not found")
	}

	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims format")
	}

	userIDStr, ok := jwtClaims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID format")
	}

	return userID, nil
}

func (s *authService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	identifier := req.Username
		if identifier == "" {
			identifier = req.Email
		}

	if identifier == "" {
         return nil, errors.New("username or email is required")
    }

	user, roleName, err := s.authRepo.FindUserByEmailOrUsername(identifier)

	if err != nil {
        return nil, errors.New("invalid username or email")
    }

    if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
        return nil, errors.New("invalid password")
    }

	if !user.ISActive {
		return nil, errors.New("account is inactive, please contact admin")
	}

	permissions, err := s.authRepo.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return nil, errors.New("failed to fetch permissions")
	}

	token, err := utils.GenerateJWT(user.ID.String(), user.Username, roleName, permissions)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &model.LoginResponse{
		Token:        token,
		RefreshToken: "",
		User: model.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			FullName:    user.FullName,
			Role:        roleName,
			Permissions: permissions,
		},
	}, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	// Get token expiration time
	expiresAt, err := utils.GetTokenExpiration(token)
	if err != nil {
		return errors.New("failed to get token expiration")
	}

	// Add token to in-memory blacklist
	utils.BlacklistManager.AddToken(token, expiresAt)
	return nil
}

func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.ProfileData, error) {
	profile, err := s.authRepo.GetUserProfile(userID)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// HTTP Endpoints
func (s *authService) LoginEndpoint(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	result, err := s.Login(c.Context(), req)
	if err != nil {
		switch err.Error() {
		case "invalid username or email":
			return c.Status(401).JSON(fiber.Map{"error": err.Error()})
		case "invalid password":
			return c.Status(401).JSON(fiber.Map{"error": err.Error()})
		case "account is inactive, please contact admin":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Login failed"})
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *authService) LogoutEndpoint(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Missing Authorization Header"})
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid Authorization Header format"})
	}

	token := authHeader[7:]
	if token == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Token is empty"})
	}

	err := s.Logout(c.Context(), token)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Logout failed"})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Logout successful",
	})
}

func (s *authService) ProfileEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaimsAuth(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	profile, err := s.GetProfile(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get profile"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   profile,
	})
}
