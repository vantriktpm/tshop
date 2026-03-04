package domain

import "time"

type Shipment struct {
	ID        string
	OrderID   string
	Address   string
	Status    string
	CreatedAt time.Time
}
