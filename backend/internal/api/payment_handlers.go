// internal/api/payment_handlers.go
// Payment processing HTTP handlers

package api

import (
	"net/http"

	"tournament-planner/internal/config"
	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// HandleProcessPayment processes a payment
func HandleProcessPayment(paymentService *services.PaymentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TournamentID  string  `json:"tournament_id" binding:"required"`
			ParticipantID string  `json:"participant_id" binding:"required"`
			Amount        float64 `json:"amount" binding:"required,min=0"`
			PaymentMethod string  `json:"payment_method" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := paymentService.ProcessPayment(c.Request.Context(), req.TournamentID, req.ParticipantID, req.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payment processed successfully"})
	}
}

// HandleRefundPayment processes a refund
func HandleRefundPayment(paymentService *services.PaymentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TournamentID  string `json:"tournament_id" binding:"required"`
			ParticipantID string `json:"participant_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := paymentService.RefundPayment(c.Request.Context(), req.TournamentID, req.ParticipantID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process refund"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Refund processed successfully"})
	}
}

// HandleStripeWebhook handles Stripe webhook events
func HandleStripeWebhook(paymentService *services.PaymentService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement Stripe webhook handling
		// This would verify the webhook signature and process events
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Stripe webhook not implemented yet"})
	}
}
