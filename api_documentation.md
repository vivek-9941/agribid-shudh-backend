# API Documentation
Auto-generated API documentation for the frontend team.

## Module: Auth

### Routes
- **POST** `/api/v1/auth/otp/send` -> Handler: `h.SendOTP`
- **POST** `/api/v1/auth/otp/verify` -> Handler: `h.VerifyOTP`
- **POST** `/api/v1/auth/login` -> Handler: `h.Login`
- **POST** `/api/v1/auth/refresh` -> Handler: `h.Refresh`
- **POST** `/api/v1/auth/logout` -> Handler: `h.Logout`
- **GET** `/api/v1/auth/me` -> Handler: `h.GetMe`

### Models / Request & Response Bodies
```go
type User struct {
	ID           uuid.UUID  `json:"id"`
	Phone        string     `json:"phone"`
	Email        *string    `json:"email,omitempty"`
	PasswordHash string     `json:"-"` // never sent to client
	FullName     string     `json:"full_name"`
	Status       string     `json:"status"` // active, inactive, suspended
	PartnerID    *uuid.UUID `json:"partner_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
```

```go
type OTPSession struct {
	ID        uuid.UUID `json:"id"`
	Phone     string    `json:"phone"`
	OTPHash   string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Verified  bool      `json:"verified"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}
```

```go
type RefreshToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}
```

```go
type UserContext struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PartnerID uuid.UUID `json:"partner_id"`
	Roles     []string  `json:"roles"`
	IsAdmin   bool      `json:"is_admin"`
}
```

```go
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expiry
}
```

```go
type SendOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
}
```

```go
type VerifyOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
	OTP   string `json:"otp" validate:"required,len=6,numeric"`
}
```

```go
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}
```

```go
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
```

```go
type RegisterRequest struct {
	Phone    string `json:"phone" validate:"required,e164"`
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password,omitempty" validate:"omitempty,min=8"`
}
```

## Module: Catalog

### Routes
- **GET** `/api/v1/products` -> Handler: `h.ListProducts`
- **POST** `/api/v1/products` -> Handler: `h.CreateProduct`
- **GET** `/api/v1/products/{id}` -> Handler: `h.GetProduct`
- **PUT** `/api/v1/products/{id}` -> Handler: `h.UpdateProduct`
- **PATCH** `/api/v1/products/{id}/status` -> Handler: `h.SetProductStatus`
- **GET** `/api/v1/products` -> Handler: `h.ListCategories`
- **POST** `/api/v1/products` -> Handler: `h.CreateCategory`

### Models / Request & Response Bodies
```go
type Product struct {
	ID             uuid.UUID              `json:"id"`
	SKU            string                 `json:"sku"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description,omitempty"`
	CategoryID     *uuid.UUID             `json:"category_id,omitempty"`
	Brand          *string                `json:"brand,omitempty"`
	ManufacturerID uuid.UUID              `json:"manufacturer_id"`
	HSNCode        string                 `json:"hsn_code"`
	UOM            *string                `json:"uom,omitempty"`
	WeightGrams    *int                   `json:"weight_grams,omitempty"`
	Dimensions     map[string]interface{}
```

```go
type ProductImage struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	URL       string    `json:"url"`
	SortOrder int       `json:"sort_order"`
	IsPrimary bool      `json:"is_primary"`
}
```

```go
type Category struct {
	ID        uuid.UUID   `json:"id"`
	ParentID  *uuid.UUID  `json:"parent_id,omitempty"`
	Name      string      `json:"name"`
	Slug      string      `json:"slug"`
	Path      string      `json:"path"`
	Level     int         `json:"level"`
	Children  []*Category `json:"children,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}
```

```go
type CreateProductRequest struct {
	SKU          string                 `json:"sku" validate:"required,min=1,max=100"`
	Name         string                 `json:"name" validate:"required,min=2,max=255"`
	Description  string                 `json:"description,omitempty"`
	CategoryID   string                 `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Brand        string                 `json:"brand,omitempty"`
	HSNCode      string                 `json:"hsn_code" validate:"required,max=20"`
	UOM          string                 `json:"uom,omitempty"`
	WeightGrams  int                    `json:"weight_grams,omitempty"`
	Dimensions   map[string]interface{}
```

```go
type UpdateProductRequest struct {
	Name         string                 `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description  string                 `json:"description,omitempty"`
	CategoryID   string                 `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Brand        string                 `json:"brand,omitempty"`
	HSNCode      string                 `json:"hsn_code,omitempty" validate:"omitempty,max=20"`
	UOM          string                 `json:"uom,omitempty"`
	WeightGrams  int                    `json:"weight_grams,omitempty"`
	Dimensions   map[string]interface{}
```

```go
type SetStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive discontinued"`
}
```

```go
type CreateCategoryRequest struct {
	ParentID string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Slug     string `json:"slug" validate:"required,min=2,max=255"`
}
```

```go
type ProductFilter struct {
	CategoryID     *uuid.UUID
	Brand          string
	Status         string
	Search         string
	ManufacturerID *uuid.UUID
	Page           int
	PageSize       int
}
```

## Module: Dashboard

### Routes
- **GET** `/dashboard` -> Handler: `h.getMetrics`
- **GET** `/dashboard/{type}` -> Handler: `h.getReport`

## Module: Dispatch

### Routes
- **POST** `/shipments` -> Handler: `h.assignDeliveryPartner`
- **PATCH** `/shipments/{id}/status` -> Handler: `h.updateStatus`
- **POST** `/shipments/{id}/pod` -> Handler: `h.recordPOD`

### Models / Request & Response Bodies
```go
type Shipment struct {
	ID                uuid.UUID      `json:"id"`
	OrderID           uuid.UUID      `json:"order_id"`
	SellerID          uuid.UUID      `json:"seller_id"`
	DeliveryPartnerID *uuid.UUID     `json:"delivery_partner_id,omitempty"`
	AssignedBy        *uuid.UUID     `json:"assigned_by,omitempty"`
	TrackingNumber    string         `json:"tracking_number"`
	Status            ShipmentStatus `json:"status"`
	VehicleNumber     *string        `json:"vehicle_number,omitempty"`
	DriverName        *string        `json:"driver_name,omitempty"`
	DriverPhone       *string        `json:"driver_phone,omitempty"`
	EstimatedDelivery *time.Time     `json:"estimated_delivery,omitempty"`
	ShippedAt         *time.Time     `json:"shipped_at,omitempty"`
	DeliveredAt       *time.Time     `json:"delivered_at,omitempty"`
	PODURL            *string        `json:"pod_url,omitempty"`
	PODType           *string        `json:"pod_type,omitempty"`
	GPSLat            *float64       `json:"gps_lat,omitempty"`
	GPSLng            *float64       `json:"gps_lng,omitempty"`
	AssignedAt        time.Time      `json:"assigned_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}
```

```go
type ShipmentEvent struct {
	ID         uuid.UUID      `json:"id"`
	ShipmentID uuid.UUID      `json:"shipment_id"`
	Status     ShipmentStatus `json:"status"`
	Location   *string        `json:"location,omitempty"`
	Notes      *string        `json:"notes,omitempty"`
	GPSLat     *float64       `json:"gps_lat,omitempty"`
	GPSLng     *float64       `json:"gps_lng,omitempty"`
	RecordedBy *uuid.UUID     `json:"recorded_by,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}
```

```go
type AssignPartnerRequest struct {
	OrderID           string `json:"order_id" validate:"required,uuid"`
	DeliveryPartnerID string `json:"delivery_partner_id" validate:"required,uuid"`
	VehicleNumber     string `json:"vehicle_number,omitempty"`
	DriverName        string `json:"driver_name,omitempty"`
	DriverPhone       string `json:"driver_phone,omitempty"`
}
```

```go
type UpdateStatusRequest struct {
	Status   string  `json:"status" validate:"required"`
	Location *string `json:"location,omitempty"`
	Notes    *string `json:"notes,omitempty"`
	GPSLat   *float64 `json:"gps_lat,omitempty"`
	GPSLng   *float64 `json:"gps_lng,omitempty"`
}
```

```go
type RecordPODRequest struct {
	PODURL  string `json:"pod_url" validate:"required,url"`
	PODType string `json:"pod_type" validate:"required,oneof=signature photo"`
}
```

## Module: Fulfillment

### Routes
- **PATCH** `/fulfillment/orders/{id}/status` -> Handler: `h.updateStatus`
- **POST** `/fulfillment/orders/{id}/partial` -> Handler: `h.recordPartial`

### Models / Request & Response Bodies
```go
type FulfillmentEvent struct {
	ID         uuid.UUID `json:"id"`
	OrderID    uuid.UUID `json:"order_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	ChangedBy  uuid.UUID `json:"changed_by"`
	Reason     *string   `json:"reason,omitempty"`
	ChangedAt  time.Time `json:"changed_at"`
}
```

```go
type PartialFulfillment struct {
	OrderLineID  uuid.UUID `json:"order_line_id"`
	FulfilledQty int       `json:"fulfilled_qty"`
}
```

```go
type RecordPartialRequest struct {
	Lines []PartialLineInput `json:"lines" validate:"required,min=1,dive"`
}
```

```go
type PartialLineInput struct {
	OrderLineID  string `json:"order_line_id" validate:"required,uuid"`
	FulfilledQty int    `json:"fulfilled_qty" validate:"required,min=1"`
}
```

## Module: Inventory

### Routes
- **GET** `/api/v1/inventory` -> Handler: `h.ListStock`
- **POST** `/api/v1/inventory/inward` -> Handler: `h.AddStock`
- **POST** `/api/v1/inventory/adjust` -> Handler: `h.AdjustStock`
- **GET** `/api/v1/inventory/low-stock` -> Handler: `h.GetLowStock`

### Models / Request & Response Bodies
```go
type InventoryLedger struct {
	ID           uuid.UUID `json:"id"`
	PartnerID    uuid.UUID `json:"partner_id"`
	WarehouseID  uuid.UUID `json:"warehouse_id"`
	ProductID    uuid.UUID `json:"product_id"`
	AvailableQty int       `json:"available_qty"`
	ReservedQty  int       `json:"reserved_qty"`
	PhysicalQty  int       `json:"physical_qty"`
	ReorderLevel int       `json:"reorder_level"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

```go
type InventoryMovement struct {
	ID            uuid.UUID  `json:"id"`
	PartnerID     uuid.UUID  `json:"partner_id"`
	WarehouseID   uuid.UUID  `json:"warehouse_id"`
	ProductID     uuid.UUID  `json:"product_id"`
	MovementType  string     `json:"movement_type"` // inward, outward, reservation, unreservation, adjustment, transfer
	Quantity      int        `json:"quantity"`
	ReferenceType string     `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	Reason        *string    `json:"reason,omitempty"`
	PerformedBy   *uuid.UUID `json:"performed_by,omitempty"`
	MovedAt       time.Time  `json:"moved_at"`
}
```

```go
type AdjustStockRequest struct {
	WarehouseID string `json:"warehouse_id" validate:"required,uuid"`
	ProductID   string `json:"product_id" validate:"required,uuid"`
	Quantity    int    `json:"quantity" validate:"required"`
	Reason      string `json:"reason" validate:"required,min=2"`
}
```

```go
type TransferStockRequest struct {
	FromWarehouseID string `json:"from_warehouse_id" validate:"required,uuid"`
	ToWarehouseID   string `json:"to_warehouse_id" validate:"required,uuid"`
	ProductID       string `json:"product_id" validate:"required,uuid"`
	Quantity        int    `json:"quantity" validate:"required,min=1"`
}
```

```go
type CreateWarehouseRequest struct {
	Name      string                 `json:"name" validate:"required,min=2,max=255"`
	Address   map[string]interface{}
```

## Module: Invoice

### Routes
- **GET** `/invoices/{id}` -> Handler: `h.getInvoice`
- **POST** `/invoices/order/{orderId}` -> Handler: `h.generateForOrder`

### Models / Request & Response Bodies
```go
type Invoice struct {
	ID              uuid.UUID      `json:"id"`
	InvoiceNumber   string         `json:"invoice_number"`
	OrderID         uuid.UUID      `json:"order_id"`
	SellerID        uuid.UUID      `json:"seller_id"`
	BuyerID         uuid.UUID      `json:"buyer_id"`
	InvoiceType     string         `json:"invoice_type"` // tax_invoice, credit_note, debit_note
	ParentInvoiceID *uuid.UUID     `json:"parent_invoice_id,omitempty"`
	SellerGSTIN     string         `json:"seller_gstin"`
	BuyerGSTIN      string         `json:"buyer_gstin"`
	PlaceOfSupply   string         `json:"place_of_supply"`
	Subtotal        string         `json:"subtotal"`
	DiscountTotal   string         `json:"discount_total"`
	TaxableAmount   string         `json:"taxable_amount"`
	CGSTTotal       string         `json:"cgst_total"`
	SGSTTotal       string         `json:"sgst_total"`
	IGSTTotal       string         `json:"igst_total"`
	CessTotal       string         `json:"cess_total"`
	GrandTotal      string         `json:"grand_total"`
	PDFURL          *string        `json:"pdf_url,omitempty"`
	Status          string         `json:"status"` // draft, issued, cancelled
	IssuedAt        *time.Time     `json:"issued_at,omitempty"`
	DueDate         *time.Time     `json:"due_date,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	Lines           []*InvoiceLine `json:"lines,omitempty"`
}
```

```go
type InvoiceLine struct {
	ID             uuid.UUID `json:"id"`
	InvoiceID      uuid.UUID `json:"invoice_id"`
	ProductID      uuid.UUID `json:"product_id"`
	SKU            string    `json:"sku"`
	Description    string    `json:"description"`
	HSNCode        string    `json:"hsn_code"`
	Quantity       int       `json:"quantity"`
	UnitPrice      string    `json:"unit_price"`
	DiscountAmount string    `json:"discount_amount"`
	TaxableAmount  string    `json:"taxable_amount"`
	CGSTRate       string    `json:"cgst_rate"`
	CGSTAmount     string    `json:"cgst_amount"`
	SGSTRate       string    `json:"sgst_rate"`
	SGSTAmount     string    `json:"sgst_amount"`
	IGSTRate       string    `json:"igst_rate"`
	IGSTAmount     string    `json:"igst_amount"`
	LineTotal      string    `json:"line_total"`
}
```

```go
type TaxBreakdown struct {
	TaxableAmount string  `json:"taxable_amount"`
	CGSTRate      string  `json:"cgst_rate"`
	CGSTAmount    string  `json:"cgst_amount"`
	SGSTRate      string  `json:"sgst_rate"`
	SGSTAmount    string  `json:"sgst_amount"`
	IGSTRate      string  `json:"igst_rate"`
	IGSTAmount    string  `json:"igst_amount"`
	CessRate      string  `json:"cess_rate"`
	CessAmount    string  `json:"cess_amount"`
	TotalTax      string  `json:"total_tax"`
}
```

```go
type HSNTaxRate struct {
	HSNCode       string `json:"hsn_code"`
	Description   string `json:"description"`
	CGSTRate      string `json:"cgst_rate"`
	SGSTRate      string `json:"sgst_rate"`
	IGSTRate      string `json:"igst_rate"`
	CessRate      string `json:"cess_rate"`
	EffectiveFrom string `json:"effective_from"`
}
```

## Module: Notification

### Routes
- **GET** `/notifications` -> Handler: `h.listNotifications`
- **PATCH** `/notifications/{id}/read` -> Handler: `h.markRead`
- **GET** `/notifications` -> Handler: `h.getPreferences`
- **PUT** `/notifications` -> Handler: `h.upsertPreferences`

### Models / Request & Response Bodies
```go
type Notification struct {
	ID          uuid.UUID  `json:"id"`
	RecipientID uuid.UUID  `json:"recipient_id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	DataPayload []byte     `json:"data_payload,omitempty"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   time.Time  `json:"created_at"`
}
```

```go
type NotificationPreference struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	EmailEnabled bool      `json:"email_enabled"`
	SMSEnabled   bool      `json:"sms_enabled"`
	InAppEnabled bool      `json:"in_app_enabled"`
	MutedEvents  []string  `json:"muted_events,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

```go
type NotificationTemplate struct {
	EventType string `json:"event_type"`
	Channel   string `json:"channel"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}
```

```go
type UpdatePreferencesRequest struct {
	Preferences []PreferenceInput `json:"preferences" validate:"required,min=1,dive"`
}
```

```go
type PreferenceInput struct {
	EventType string `json:"event_type" validate:"required"`
	Channel   string `json:"channel" validate:"required,oneof=in_app sms email push"`
	Enabled   bool   `json:"enabled"`
}
```

## Module: Order

### Routes
- **POST** `/cart` -> Handler: `h.createCart`
- **GET** `/cart` -> Handler: `h.getCart`
- **DELETE** `/cart` -> Handler: `h.deleteCart`
- **POST** `/cart` -> Handler: `h.addCartItem`
- **PUT** `/cart/{id}` -> Handler: `h.updateCartItem`
- **DELETE** `/cart/{id}` -> Handler: `h.removeCartItem`
- **POST** `/cart` -> Handler: `h.placeOrder`
- **GET** `/cart` -> Handler: `h.listOrders`
- **GET** `/cart/{id}` -> Handler: `h.getOrder`
- **DELETE** `/cart/{id}` -> Handler: `h.cancelOrder`

### Models / Request & Response Bodies
```go
type Cart struct {
	ID        uuid.UUID   `json:"id"`
	BuyerID   uuid.UUID   `json:"buyer_id"`
	SellerID  uuid.UUID   `json:"seller_id"`
	Status    string      `json:"status"` // active, ordered, abandoned
	Items     []*CartItem `json:"items,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
```

```go
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
```

```go
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
	DeliveryAddress map[string]interface{}
```

```go
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
```

```go
type OrderStatusHistory struct {
	ID         uuid.UUID  `json:"id"`
	OrderID    uuid.UUID  `json:"order_id"`
	FromStatus string     `json:"from_status"`
	ToStatus   string     `json:"to_status"`
	ChangedBy  *uuid.UUID `json:"changed_by,omitempty"`
	Reason     *string    `json:"reason,omitempty"`
	ChangedAt  time.Time  `json:"changed_at"`
}
```

```go
type CreateCartRequest struct {
	SellerID string `json:"seller_id" validate:"required,uuid"`
}
```

```go
type AddCartItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}
```

```go
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}
```

```go
type PlaceOrderRequest struct {
	DeliveryAddress map[string]interface{}
```

```go
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=confirmed packed shipped delivered cancelled"`
	Reason string `json:"reason,omitempty"`
}
```

## Module: Partner

### Routes
- **POST** `/api/v1/partners` -> Handler: `h.Create`
- **GET** `/api/v1/partners` -> Handler: `h.List`
- **GET** `/api/v1/partners/{id}` -> Handler: `h.GetByID`
- **PUT** `/api/v1/partners/{id}` -> Handler: `h.Update`
- **POST** `/api/v1/partners/{id}/kyc` -> Handler: `h.SubmitKYC`
- **GET** `/api/v1/partners/{id}/kyc` -> Handler: `h.GetKYC`
- **PATCH** `/api/v1/partners/{id}/kyc` -> Handler: `h.ReviewKYC`
- **GET** `/api/v1/partners/{id}/children` -> Handler: `h.GetChildren`

### Models / Request & Response Bodies
```go
type Partner struct {
	ID           uuid.UUID  `json:"id"`
	Code         string     `json:"code"`
	Type         string     `json:"type"` // manufacturer, state_stockist, distributor, sub_distributor, retailer, delivery_partner
	BusinessName string     `json:"business_name"`
	TradeName    *string    `json:"trade_name,omitempty"`
	ParentID     *uuid.UUID `json:"parent_id,omitempty"`
	GSTIN        *string    `json:"gstin,omitempty"`
	PAN          *string    `json:"pan,omitempty"`
	StateCode    *string    `json:"state_code,omitempty"`
	Address      *Address   `json:"address,omitempty"`
	Status       string     `json:"status"` // pending, active, inactive, suspended
	KYCStatus    string     `json:"kyc_status"` // pending, submitted, approved, rejected
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
```

```go
type Address struct {
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	Pincode string `json:"pincode"`
}
```

```go
type KYCDocument struct {
	ID          uuid.UUID  `json:"id"`
	PartnerID   uuid.UUID  `json:"partner_id"`
	DocType     string     `json:"doc_type"` // gst_certificate, pan_card, business_license, bank_details, address_proof
	FileURL     string     `json:"file_url"`
	Status      string     `json:"status"` // pending, approved, rejected
	ReviewedBy  *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewNote  *string    `json:"review_note,omitempty"`
	SubmittedAt time.Time  `json:"submitted_at"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
}
```

```go
type Warehouse struct {
	ID        uuid.UUID `json:"id"`
	PartnerID uuid.UUID `json:"partner_id"`
	Name      string    `json:"name"`
	Address   *Address  `json:"address,omitempty"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}
```

```go
type CreatePartnerRequest struct {
	Type         string   `json:"type" validate:"required,oneof=manufacturer state_stockist distributor sub_distributor retailer delivery_partner"`
	BusinessName string   `json:"business_name" validate:"required,min=2,max=255"`
	TradeName    string   `json:"trade_name,omitempty"`
	ParentID     string   `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	GSTIN        string   `json:"gstin,omitempty" validate:"omitempty,len=15"`
	PAN          string   `json:"pan,omitempty" validate:"omitempty,len=10"`
	StateCode    string   `json:"state_code,omitempty" validate:"omitempty,len=2"`
	Address      *Address `json:"address,omitempty"`
}
```

```go
type UpdatePartnerRequest struct {
	BusinessName string   `json:"business_name,omitempty" validate:"omitempty,min=2,max=255"`
	TradeName    string   `json:"trade_name,omitempty"`
	GSTIN        string   `json:"gstin,omitempty" validate:"omitempty,len=15"`
	PAN          string   `json:"pan,omitempty" validate:"omitempty,len=10"`
	StateCode    string   `json:"state_code,omitempty" validate:"omitempty,len=2"`
	Address      *Address `json:"address,omitempty"`
}
```

```go
type SubmitKYCRequest struct {
	DocType string `json:"doc_type" validate:"required,oneof=gst_certificate pan_card business_license bank_details address_proof"`
	FileURL string `json:"file_url" validate:"required,url"`
}
```

```go
type ReviewKYCRequest struct {
	Status     string `json:"status" validate:"required,oneof=approved rejected"`
	ReviewNote string `json:"review_note,omitempty"`
}
```

## Module: Payment

### Routes
- **POST** `/payments` -> Handler: `h.recordPayment`
- **GET** `/payments/ledger` -> Handler: `h.getLedger`
- **GET** `/payments/aging` -> Handler: `h.getAgingReport`
- **GET** `/payments/credit-status` -> Handler: `h.getCreditStatus`

### Models / Request & Response Bodies
```go
type Payment struct {
	ID              uuid.UUID  `json:"id"`
	PaymentNumber   string     `json:"payment_number"`
	BuyerID         uuid.UUID  `json:"buyer_id"`
	SellerID        uuid.UUID  `json:"seller_id"`
	Amount          string     `json:"amount"`
	Method          string     `json:"method"` // cash, bank_transfer, cheque, upi, card, gateway
	ReferenceNumber *string    `json:"reference_number,omitempty"`
	GatewayTxnID    *string    `json:"gateway_txn_id,omitempty"`
	InvoiceID       *uuid.UUID `json:"invoice_id,omitempty"`
	Status          string     `json:"status"` // pending, confirmed, failed, refunded
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
```

```go
type CreditAccount struct {
	ID             uuid.UUID `json:"id"`
	BuyerID        uuid.UUID `json:"buyer_id"`
	SellerID       uuid.UUID `json:"seller_id"`
	CreditLimit    string    `json:"credit_limit"`
	CreditUtilized string    `json:"credit_utilized"`
	CreditAvailable string   `json:"credit_available"` // computed
	PaymentTerms   int       `json:"payment_terms"`
	OverdueAmount  string    `json:"overdue_amount"`
	UpdatedAt      time.Time `json:"updated_at"`
}
```

```go
type LedgerEntry struct {
	ID             uuid.UUID  `json:"id"`
	PartnerID      uuid.UUID  `json:"partner_id"`
	CounterpartyID uuid.UUID  `json:"counterparty_id"`
	EntryType      string     `json:"entry_type"` // invoice, payment, credit_note, adjustment
	ReferenceID    *uuid.UUID `json:"reference_id,omitempty"`
	Debit          string     `json:"debit"`
	Credit         string     `json:"credit"`
	Balance        string     `json:"balance"`
	EntryDate      time.Time  `json:"entry_date"`
	Description    string     `json:"description"`
	CreatedAt      time.Time  `json:"created_at"`
}
```

```go
type AgingBucket struct {
	Bucket   string `json:"bucket"` // 0-30, 31-60, 61-90, 90+
	Amount   string `json:"amount"`
	Count    int    `json:"count"`
}
```

```go
type RecordPaymentRequest struct {
	SellerID        string `json:"seller_id" validate:"required,uuid"`
	Amount          string `json:"amount" validate:"required"`
	Method          string `json:"method" validate:"required,oneof=cash bank_transfer cheque upi card gateway"`
	ReferenceNumber string `json:"reference_number,omitempty"`
	InvoiceID       string `json:"invoice_id,omitempty" validate:"omitempty,uuid"`
}
```

## Module: Pricing

### Routes
- **POST** `/schemes/products/{id}/prices` -> Handler: `h.setProductPrice`
- **POST** `/schemes` -> Handler: `h.createScheme`
- **PUT** `/schemes/{id}` -> Handler: `h.updateScheme`
- **PATCH** `/schemes/{id}/status` -> Handler: `h.setSchemeStatus`
- **GET** `/schemes/products/{id}/calculate-price` -> Handler: `h.calculatePrice`

### Models / Request & Response Bodies
```go
type ProductPrice struct {
	ID            uuid.UUID  `json:"id"`
	ProductID     uuid.UUID  `json:"product_id"`
	RoleCode      string     `json:"role_code"`
	PartnerID     *uuid.UUID `json:"partner_id,omitempty"`
	MRP           string     `json:"mrp"` // string to avoid float precision
	BasePrice     string     `json:"base_price"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
```

```go
type Scheme struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"` // percentage, flat, buy_x_get_y
	DiscountValue string    `json:"discount_value,omitempty"`
	BuyQty        *int      `json:"buy_qty,omitempty"`
	GetQty        *int      `json:"get_qty,omitempty"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidTo       time.Time `json:"valid_to"`
	IsExclusive   bool      `json:"is_exclusive"`
	CreatedBy     uuid.UUID `json:"created_by"`
	Status        string    `json:"status"` // active, inactive
	CreatedAt     time.Time `json:"created_at"`
}
```

```go
type SchemeApplicability struct {
	ID         uuid.UUID `json:"id"`
	SchemeID   uuid.UUID `json:"scheme_id"`
	TargetType string    `json:"target_type"` // role, partner, category, product
	TargetID   string    `json:"target_id"`
}
```

```go
type PriceResult struct {
	BasePrice     string     `json:"base_price"`
	DiscountAmount string    `json:"discount_amount"`
	TaxableAmount string     `json:"taxable_amount"`
	CGSTRate      string     `json:"cgst_rate"`
	CGSTAmount    string     `json:"cgst_amount"`
	SGSTRate      string     `json:"sgst_rate"`
	SGSTAmount    string     `json:"sgst_amount"`
	IGSTRate      string     `json:"igst_rate"`
	IGSTAmount    string     `json:"igst_amount"`
	LineTotal     string     `json:"line_total"`
	AppliedSchemes []uuid.UUID `json:"applied_schemes,omitempty"`
}
```

```go
type SetPriceRequest struct {
	RoleCode      string `json:"role_code" validate:"required"`
	PartnerID     string `json:"partner_id,omitempty" validate:"omitempty,uuid"`
	MRP           string `json:"mrp" validate:"required"`
	BasePrice     string `json:"base_price" validate:"required"`
	EffectiveFrom string `json:"effective_from" validate:"required"`
	EffectiveTo   string `json:"effective_to,omitempty"`
}
```

```go
type CreateSchemeRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Type          string `json:"type" validate:"required,oneof=percentage flat buy_x_get_y"`
	DiscountValue string `json:"discount_value,omitempty"`
	BuyQty        *int   `json:"buy_qty,omitempty"`
	GetQty        *int   `json:"get_qty,omitempty"`
	ValidFrom     string `json:"valid_from" validate:"required"`
	ValidTo       string `json:"valid_to" validate:"required"`
	IsExclusive   bool   `json:"is_exclusive"`
	Targets       []SchemeTarget `json:"targets,omitempty"`
}
```

```go
type SchemeTarget struct {
	TargetType string `json:"target_type" validate:"required,oneof=role partner category product"`
	TargetID   string `json:"target_id" validate:"required"`
}
```

```go
type UpdateSchemeRequest struct {
	Name          string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	DiscountValue string `json:"discount_value,omitempty"`
	ValidFrom     string `json:"valid_from,omitempty"`
	ValidTo       string `json:"valid_to,omitempty"`
	IsExclusive   *bool  `json:"is_exclusive,omitempty"`
}
```

```go
type SetSchemeStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive"`
}
```

## Module: Rbac

### Models / Request & Response Bodies
```go
type Role struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
```

```go
type Permission struct {
	ID       uuid.UUID `json:"id"`
	RoleID   uuid.UUID `json:"role_id"`
	Resource string    `json:"resource"`
	Action   string    `json:"action"`
}
```

## Module: Returns

### Routes
- **POST** `/returns` -> Handler: `h.initiateReturn`
- **PATCH** `/returns/{id}/status` -> Handler: `h.updateStatus`
- **POST** `/returns/{id}/inspection` -> Handler: `h.recordInspection`

### Models / Request & Response Bodies
```go
type ReturnRequest struct {
	ID                uuid.UUID     `json:"id"`
	ReturnNumber      string        `json:"return_number"`
	OrderID           uuid.UUID     `json:"order_id"`
	InvoiceID         *uuid.UUID    `json:"invoice_id,omitempty"`
	BuyerID           uuid.UUID     `json:"buyer_id"`
	SellerID          uuid.UUID     `json:"seller_id"`
	Status            ReturnStatus  `json:"status"`
	Reason            string        `json:"reason"` // damaged, expired, wrong_product, quality_issue, other
	Notes             *string       `json:"notes,omitempty"`
	PickupScheduledAt *time.Time    `json:"pickup_scheduled_at,omitempty"`
	CreditNoteID      *uuid.UUID    `json:"credit_note_id,omitempty"`
	Lines             []*ReturnLine `json:"lines,omitempty"`
	RequestedAt       time.Time     `json:"requested_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}
```

```go
type ReturnLine struct {
	ID               uuid.UUID `json:"id"`
	ReturnID         uuid.UUID `json:"return_id"`
	OrderLineID      uuid.UUID `json:"order_line_id"`
	ProductID        uuid.UUID `json:"product_id"`
	RequestedQty     int       `json:"requested_qty"`
	ApprovedQty      *int      `json:"approved_qty,omitempty"`
	InspectionResult *string   `json:"inspection_result,omitempty"` // pass, fail, partial
	InspectionNotes  *string   `json:"inspection_notes,omitempty"`
}
```

```go
type InitiateReturnRequest struct {
	OrderID string            `json:"order_id" validate:"required,uuid"`
	Reason  string            `json:"reason" validate:"required,oneof=damaged expired wrong_product quality_issue other"`
	Notes   string            `json:"notes,omitempty"`
	Lines   []ReturnLineInput `json:"lines" validate:"required,min=1,dive"`
}
```

```go
type ReturnLineInput struct {
	OrderLineID  string `json:"order_line_id" validate:"required,uuid"`
	RequestedQty int    `json:"requested_qty" validate:"required,min=1"`
}
```

```go
type InspectionInput struct {
	Lines []InspectionLineInput `json:"lines" validate:"required,min=1,dive"`
}
```

```go
type InspectionLineInput struct {
	ReturnLineID    string `json:"return_line_id" validate:"required,uuid"`
	Result          string `json:"result" validate:"required,oneof=pass fail partial"`
	InspectionNotes string `json:"inspection_notes,omitempty"`
}
```

