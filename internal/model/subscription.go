package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name"`
	MonthlyCost int        `json:"monthly_cost" db:"monthly_cost"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// JSON методы для кастомного форматирования дат
func (s Subscription) MarshalJSON() ([]byte, error) {
	type Alias Subscription
	return json.Marshal(&struct {
		StartDate string  `json:"start_date"`
		EndDate   *string `json:"end_date,omitempty"`
		CreatedAt string  `json:"created_at"`
		UpdatedAt string  `json:"updated_at"`
		*Alias
	}{
		StartDate: formatMonthYear(s.StartDate),
		EndDate:   formatMonthYearPtr(s.EndDate),
		CreatedAt: formatDateTime(s.CreatedAt),
		UpdatedAt: formatDateTime(s.UpdatedAt),
		Alias:     (*Alias)(&s),
	})
}

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" binding:"required"`
	MonthlyCost int       `json:"monthly_cost" binding:"required,min=1"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	StartDate   string    `json:"start_date" binding:"required"`
	EndDate     *string   `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" binding:"required"`
	MonthlyCost int       `json:"monthly_cost" binding:"required,min=1"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	StartDate   string    `json:"start_date" binding:"required"`
	EndDate     *string   `json:"end_date,omitempty"`
}

type SummaryFilter struct {
	UserID      uuid.UUID `form:"user_id"`
	ServiceName string    `form:"service_name"`
	StartPeriod string    `form:"start_period" binding:"required"`
	EndPeriod   string    `form:"end_period" binding:"required"`
}

type SummaryResponse struct {
	TotalCost int `json:"total_cost"`
}

// Вспомогательные функции для форматирования дат
func formatMonthYear(t time.Time) string {
	// Формат "01-2006" (месяц-год)
	return t.Format("01-2006")
}

func formatMonthYearPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := formatMonthYear(*t)
	return &formatted
}

func formatDateTime(t time.Time) string {
	// Дата и время для created_at/updated_at
	location, _ := time.LoadLocation("Europe/Moscow")
	return t.In(location).Format("2006-01-02 15:04:05")
}

// Функции для парсинга периодов (месяц-год)
func ParseMonthYear(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date string is empty")
	}
	// Парсим в формате "01-2006" (месяц-год)
	// Устанавливаем день = 1 (первое число месяца)
	return time.Parse("01-2006", dateStr)
}

func ParseMonthYearPtr(dateStr *string) (*time.Time, error) {
	if dateStr == nil || *dateStr == "" {
		return nil, nil
	}
	t, err := time.Parse("01-2006", *dateStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
