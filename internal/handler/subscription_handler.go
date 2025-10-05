package handler

import (
	"net/http"

	"github.com/Zipklas/subscription-service/internal/logger"
	"github.com/Zipklas/subscription-service/internal/model"
	"github.com/Zipklas/subscription-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service service.SubscriptionService
	logger  *logger.Logger
}

func NewSubscriptionHandler(service service.SubscriptionService, logger *logger.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscription создает новую подписку
// @Summary Создать подписку
// @Description Создает новую запись о подписке пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body model.CreateSubscriptionRequest true "Данные для создания подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req model.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid request body for subscription creation",
			"error", err,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "Creating new subscription",
		"service_name", req.ServiceName,
		"user_id", req.UserID,
	)

	subscription, err := h.service.CreateSubscription(c.Request.Context(), req)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to create subscription",
			"service_name", req.ServiceName,
			"user_id", req.UserID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "Subscription created successfully",
		"subscription_id", subscription.ID,
		"service_name", req.ServiceName,
	)

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription получает подписку по ID
// @Summary Получить подписку
// @Description Возвращает информацию о подписке по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid subscription ID format",
			"subscription_id", c.Param("id"),
			"error", err,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	h.logger.Debug(c.Request.Context(), "Getting subscription",
		"subscription_id", id,
	)

	subscription, err := h.service.GetSubscription(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "subscription not found" {
			h.logger.Warn(c.Request.Context(), "Subscription not found",
				"subscription_id", id,
			)
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		h.logger.Error(c.Request.Context(), "Failed to get subscription",
			"subscription_id", id,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Debug(c.Request.Context(), "Subscription retrieved successfully",
		"subscription_id", id,
	)

	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription обновляет подписку
// @Summary Обновить подписку
// @Description Обновляет информацию о подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param request body model.UpdateSubscriptionRequest true "Данные для обновления подписки"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid subscription ID format for update",
			"subscription_id", c.Param("id"),
			"error", err,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	var req model.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid request body for subscription update",
			"subscription_id", id,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "Updating subscription",
		"subscription_id", id,
		"service_name", req.ServiceName,
		"user_id", req.UserID,
	)

	if err := h.service.UpdateSubscription(c.Request.Context(), id, req); err != nil {
		if err.Error() == "subscription not found" {
			h.logger.Warn(c.Request.Context(), "Subscription not found for update",
				"subscription_id", id,
			)
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		h.logger.Error(c.Request.Context(), "Failed to update subscription",
			"subscription_id", id,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "Subscription updated successfully",
		"subscription_id", id,
	)

	c.JSON(http.StatusOK, SuccessResponse{Message: "subscription updated successfully"})
}

// DeleteSubscription удаляет подписку
// @Summary Удалить подписку
// @Description Удаляет запись о подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn(c.Request.Context(), "Invalid subscription ID format for deletion",
			"subscription_id", c.Param("id"),
			"error", err,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	h.logger.Info(c.Request.Context(), "Deleting subscription",
		"subscription_id", id,
	)

	if err := h.service.DeleteSubscription(c.Request.Context(), id); err != nil {
		if err.Error() == "subscription not found" {
			h.logger.Warn(c.Request.Context(), "Subscription not found for deletion",
				"subscription_id", id,
			)
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		h.logger.Error(c.Request.Context(), "Failed to delete subscription",
			"subscription_id", id,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "Subscription deleted successfully",
		"subscription_id", id,
	)

	c.JSON(http.StatusOK, SuccessResponse{Message: "subscription deleted successfully"})
}

// ListSubscriptions возвращает список подписок
// @Summary Список подписок
// @Description Возвращает список подписок с возможностью фильтрации по пользователю и сервису
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string false "ID пользователя для фильтрации"
// @Param service_name query string false "Название сервиса для фильтрации"
// @Success 200 {array} model.Subscription
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	var userID *uuid.UUID
	var serviceName *string

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			userID = &id
		}
	}

	if serviceNameStr := c.Query("service_name"); serviceNameStr != "" {
		serviceName = &serviceNameStr
	}

	h.logger.Debug(c.Request.Context(), "Listing subscriptions",
		"user_id", userID,
		"service_name", serviceName,
	)

	subscriptions, err := h.service.ListSubscriptions(c.Request.Context(), userID, serviceName)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to list subscriptions",
			"user_id", userID,
			"service_name", serviceName,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Debug(c.Request.Context(), "Subscriptions listed successfully",
		"count", len(subscriptions),
		"user_id", userID,
	)

	c.JSON(http.StatusOK, subscriptions)
}

// CalculateTotalCost подсчитывает суммарную стоимость подписок
// @Summary Подсчет стоимости
// @Description Подсчитывает суммарную стоимость всех подписок за выбранный период с фильтрацией
// @Tags summary
// @Accept json
// @Produce json
// @Param user_id query string false "ID пользователя для фильтрации"
// @Param service_name query string false "Название сервиса для фильтрации"
// @Param start_period query string true "Начало периода (формат: MM-YYYY)"
// @Param end_period query string true "Конец периода (формат: MM-YYYY)"
// @Success 200 {object} model.SummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/summary [get]
func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	var filter model.SummaryFilter

	// Парсим user_id вручную, если передан
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Warn(c.Request.Context(), "Invalid user_id format",
				"user_id", userIDStr,
				"error", err,
			)
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id format"})
			return
		}
		filter.UserID = userID
	}

	// Парсим остальные параметры
	filter.ServiceName = c.Query("service_name")
	filter.StartPeriod = c.Query("start_period")
	filter.EndPeriod = c.Query("end_period")

	// Валидация обязательных полей
	if filter.StartPeriod == "" || filter.EndPeriod == "" {
		h.logger.Warn(c.Request.Context(), "Missing required parameters for cost calculation",
			"start_period", filter.StartPeriod,
			"end_period", filter.EndPeriod,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "start_period and end_period are required"})
		return
	}

	h.logger.Info(c.Request.Context(), "Calculating total cost",
		"start_period", filter.StartPeriod,
		"end_period", filter.EndPeriod,
		"user_id", filter.UserID,
		"service_name", filter.ServiceName,
	)

	result, err := h.service.CalculateTotalCost(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to calculate total cost",
			"start_period", filter.StartPeriod,
			"end_period", filter.EndPeriod,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "Total cost calculated successfully",
		"total_cost", result.TotalCost,
		"start_period", filter.StartPeriod,
		"end_period", filter.EndPeriod,
	)

	c.JSON(http.StatusOK, result)
}

// Вспомогательные структуры для ответов
type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
