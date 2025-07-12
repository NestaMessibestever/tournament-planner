// internal/api/auth_handlers.go
// Authentication-related HTTP handlers

package api

import (
	"net/http"

	"tournament-planner/internal/models"
	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// HandleRegister handles user registration
func HandleRegister(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		user, tokens, err := authService.Register(c.Request.Context(), req)
		if err != nil {
			if err == services.ErrEmailAlreadyExists {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"user": user,
			"auth": tokens,
		})
	}
}

// HandleLogin handles user login
func HandleLogin(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		user, tokens, err := authService.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			if err == services.ErrInvalidCredentials {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": user,
			"auth": tokens,
		})
	}
}

// HandleLogout handles user logout
func HandleLogout(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get refresh token from request
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		c.ShouldBindJSON(&req)

		// Invalidate refresh token
		if err := authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
			// Log error but don't fail the logout
			c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

// HandleRefreshToken handles token refresh
func HandleRefreshToken(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		tokens, err := authService.RefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			if err == services.ErrInvalidToken {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"auth": tokens,
		})
	}
}

// HandleForgotPassword handles password reset request
func HandleForgotPassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Always return success to not reveal if email exists
		authService.ForgotPassword(c.Request.Context(), req.Email)
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
	}
}

// HandleResetPassword handles password reset
func HandleResetPassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token       string `json:"token" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := authService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
			if err == services.ErrInvalidToken {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
	}
}

// HandleVerifyEmail handles email verification
func HandleVerifyEmail(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := authService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
			if err == services.ErrInvalidToken {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification token"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
	}
}

// HandleChangePassword handles password change for authenticated users
func HandleChangePassword(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			CurrentPassword string `json:"current_password" binding:"required"`
			NewPassword     string `json:"new_password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := authService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			if err == services.ErrInvalidCredentials {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}
