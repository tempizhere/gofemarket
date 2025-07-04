package model

import "time"

// Balance представляет баланс пользователя.
type Balance struct {
	Current   float64
	Withdrawn float64
}

// Withdrawal представляет списание баллов.
type Withdrawal struct {
	Order       string
	Sum         float64
	ProcessedAt time.Time
}
