package domain

import "time"

type Promotion struct {
	ID         string
	Code       string
	Discount   float64
	ValidUntil time.Time
}
