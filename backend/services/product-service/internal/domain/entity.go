package domain

import "time"

type Product struct {
	ProductID   string     // uuid NOT NULL
	ProductCode string     // character(50) NOT NULL
	ProductName string     // character(255) NOT NULL
	Quantity    float64    // numeric NOT NULL DEFAULT 0
	Price       float64    // numeric NOT NULL DEFAULT 0
	PriceSale   float64    // numeric DEFAULT 0
	CreatedBy   *string    // character(50), nullable
	UpdatedBy   *string    // character(50), nullable
	CreatedDate *time.Time // timestamp with time zone, nullable
	UpdatedDate *time.Time // timestamp with time zone, nullable
	ImageID     *string    // uuid, nullable
}
