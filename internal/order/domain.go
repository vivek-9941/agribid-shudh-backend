package order

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents valid order states.
type OrderStatus string

const (
	StatusPending            OrderStatus = "pending"
	StatusConfirmed          OrderStatus = "confirmed"
	StatusPacked             OrderStatus = "packed"
	StatusShipped            OrderStatus = "shipped"
	StatusDelivered          OrderStatus = "delivered"
	StatusCancelled          OrderStatus = "cancelled"
	StatusPartiallyFulfilled OrderStatus = "partially_fulfilled"
)

// ValidTransitions defines the state machine for order status changes.
var ValidTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:            {StatusConfirmed, StatusCancelled},
	StatusConfirmed:          {StatusPacked, StatusCancelled},
	StatusPacked:             {StatusShipped, StatusCancelled},
	StatusShipped:            {StatusDelivered, StatusCancelled, StatusPartiallyFulfilled},
	StatusPartiallyFulfilled: {StatusDelivered, StatusCancelled},
}

// CanTransition checks if a transition from→to is valid.
func CanTransition(from, to OrderStatus) bool {
	targets, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// Cart represents a shopping cart.
type Cart struct {
	ID        uuid.UUID   `json:"id"`
	BuyerID   uuid.UUID   `json:"buyer_id"`
	SellerID  uuid.UUID   `json:"seller_id"`
	Status    string      `json:"status"` // active, ordered, abandoned
	Items     []*CartItem `json:"items,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// CartItem represents a single item in a cart.
type CartItem struct {
	ID        uuid.UUID   `json:"id"`
	CartID    uuid.UUID   `json:"cart_id"`
	ProductID uuid.UUID   `json:"product_id"`
	Quantity  int         `json:"quantity"`
	UnitPrice string      `json:"unit_price"`
	SchemeIDs []uuid.UUID `json:"scheme_ids,omitempty"`
	LineTotal string      `json:"line_total"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Order represents a placed order.
type Order struct {
	ID              uuid.UUID              `json:"id"`
	OrderNumber     string                 `json:"order_number"`
	BuyerID         uuid.UUID              `json:"buyer_id"`
	SellerID        uuid.UUID              `json:"seller_id"`
	CartID          *uuid.UUID             `json:"cart_id,omitempty"`
	Status          OrderStatus            `json:"status"`
	Subtotal        string                 `json:"subtotal"`
	DiscountTotal   string                 `json:"discount_total"`
	TaxableAmount   string                 `json:"taxable_amount"`
	CGSTTotal       string                 `json:"cgst_total"`
	SGSTTotal       string                 `json:"sgst_total"`
	IGSTTotal       string                 `json:"igst_total"`
	GrandTotal      string                 `json:"grand_total"`
	PaymentStatus   string                 `json:"payment_status"`
	DeliveryAddress map[string]interface{} `json:"delivery_address,omitempty"`
	Notes           *string                `json:"notes,omitempty"`
	Lines           []*OrderLine           `json:"lines,omitempty"`
	PlacedAt        *time.Time             `json:"placed_at,omitempty"`
	ConfirmedAt     *time.Time             `json:"confirmed_at,omitempty"`
	DeliveredAt     *time.Time             `json:"delivered_at,omitempty"`
	CancelledAt     *time.Time             `json:"cancelled_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// OrderLine represents a line item in an order.
type OrderLine struct {
	ID             uuid.UUID `json:"id"`
	OrderID        uuid.UUID `json:"order_id"`
	ProductID      uuid.UUID `json:"product_id"`
	SKU            string    `json:"sku"`
	ProductName    string    `json:"product_name"`
	OrderedQty     int       `json:"ordered_qty"`
	FulfilledQty   int       `json:"fulfilled_qty"`
	UnitPrice      string    `json:"unit_price"`
	DiscountAmount string    `json:"discount_amount"`
	TaxableAmount  string    `json:"taxable_amount"`
	CGSTAmount     string    `json:"cgst_amount"`
	SGSTAmount     string    `json:"sgst_amount"`
	IGSTAmount     string    `json:"igst_amount"`
	LineTotal      string    `json:"line_total"`
	HSNCode        string    `json:"hsn_code"`
}

// OrderStatusHistory records a status change.
type OrderStatusHistory struct {
	ID         uuid.UUID  `json:"id"`
	OrderID    uuid.UUID  `json:"order_id"`
	FromStatus string     `json:"from_status"`
	ToStatus   string     `json:"to_status"`
	ChangedBy  *uuid.UUID `json:"changed_by,omitempty"`
	Reason     *string    `json:"reason,omitempty"`
	ChangedAt  time.Time  `json:"changed_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// CreateCartRequest is the payload for POST /cart.
type CreateCartRequest struct {
	SellerID string `json:"seller_id" validate:"required,uuid"`
}

// AddCartItemRequest is the payload for POST /cart/items.
type AddCartItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

// UpdateCartItemRequest is the payload for PUT /cart/items/:id.
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}

// PlaceOrderRequest is the payload for POST /orders.
type PlaceOrderRequest struct {
	DeliveryAddress map[string]interface{} `json:"delivery_address,omitempty"`
	Notes           string                 `json:"notes,omitempty"`
}

// UpdateOrderStatusRequest is the payload for PATCH /fulfillment/orders/:id/status.
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=confirmed packed shipped delivered cancelled"`
	Reason string `json:"reason,omitempty"`
}
