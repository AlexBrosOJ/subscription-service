package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"subscription-service/internal/models"
	"subscription-service/internal/service"
)

type SubscriptionHandler struct {
	service service.ISubscriptionService
	logger  *logrus.Logger
}

func NewSubscriptionHandler(service service.ISubscriptionService, logger *logrus.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscription создает новую подписку
// @Summary Создать подписку
// @Description Создает новую запись о подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body models.CreateSubscriptionRequest true "Данные подписки"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	sub, err := h.service.Create(ctx, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"subscription_id": sub.ID,
		"user_id":         sub.UserID,
	}).Info("Subscription created successfully")

	c.JSON(http.StatusCreated, sub)
}

// GetSubscription получает подписку по ID
// @Summary Получить подписку
// @Description Возвращает подписку по её ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	sub, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription обновляет подписку
// @Summary Обновить подписку
// @Description Обновляет существующую подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param request body models.UpdateSubscriptionRequest true "Данные для обновления"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	sub, err := h.service.Update(ctx, id, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	h.logger.WithField("subscription_id", id).Info("Subscription updated successfully")
	c.JSON(http.StatusOK, sub)
}

// DeleteSubscription удаляет подписку
// @Summary Удалить подписку
// @Description Удаляет подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	err := h.service.Delete(ctx, id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithField("subscription_id", id).Info("Subscription deleted successfully")
	c.Status(http.StatusNoContent)
}

// ListSubscriptions возвращает список подписок
// @Summary Список подписок
// @Description Возвращает список подписок с пагинацией
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Лимит записей" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Subscription
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	ctx := c.Request.Context()
	subscriptions, err := h.service.List(ctx, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list subscriptions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// GetTotalPrice возвращает суммарную стоимость подписок за период
// @Summary Суммарная стоимость
// @Description Рассчитывает суммарную стоимость подписок за выбранный период
// @Tags summary
// @Produce json
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string true "Дата начала (MM-YYYY)"
// @Param end_date query string true "Дата окончания (MM-YYYY)"
// @Success 200 {object} models.SummaryResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions/summary [get]
func (h *SubscriptionHandler) GetTotalPrice(c *gin.Context) {
	var req models.SummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	totalPrice, err := h.service.GetTotalPrice(ctx, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to calculate total price")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"service_name": req.ServiceName,
		"total_price":  totalPrice,
	}).Info("Total price calculated")

	c.JSON(http.StatusOK, models.SummaryResponse{TotalPrice: totalPrice})
}
