package domain

type Cart struct {
	UserID string
	Items  []CartItem
}

type CartItem struct {
	ProductID string
	Quantity  int64
}
