package model

import "time"

// Order представляет заказ.
type Order struct {
	ID         int
	Number     string
	UserID     int
	Status     string
	Accrual    float64
	UploadedAt time.Time
}
