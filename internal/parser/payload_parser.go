package parser

import (
	"encoding/json"
	"fmt"
	"time"
)

type Order struct {
	OrderID       string      `json:"id"`
	OrderNumber   string      `json:"orderNumber"`
	PackageStatus string      `json:"packageStatus"`
	CreatedAt     time.Time   `json:"orderDate"`
	Lines         []OrderLine `json:"lines"`
	ShipmentInfo  Shipment    `json:"shipmentAddress"`
	CargoProvider string      `json:"cargoProviderName"`
}

type OrderLine struct {
	ProductName string  `json:"productName"`
	Barcode     string  `json:"barcode"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"amount"`
}

type Shipment struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Address1   string `json:"address1"`
	City       string `json:"city"`
	District   string `json:"district"`
	PostalCode string `json:"postalCode"`
}

// DBRow represents the top-level structure of a row in the trendyol_orders table
type DBRow struct {
	OrderID       string          `json:"order_id"`
	OrderNumber   string          `json:"order_number"`
	PackageStatus string          `json:"package_status"`
	Payload       json.RawMessage `json:"payload"`
}

func ParseOrder(raw string) (*Order, error) {
	// First, parse the raw database row
	var row DBRow
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		return nil, fmt.Errorf("DB row JSON parse hatasi: %w", err)
	}

	// Then, parse the actual Trendyol order payload from the row's 'payload' column
	var order Order
	if err := json.Unmarshal(row.Payload, &order); err != nil {
		return nil, fmt.Errorf("Trendyol payload JSON parse hatasi: %w", err)
	}

	// Ensure the top-level DB fields match the payload fields (optional but good for integrity)
	if order.OrderID == "" {
		order.OrderID = row.OrderID
	}
	if order.OrderNumber == "" {
		order.OrderNumber = row.OrderNumber
	}
	if order.PackageStatus == "" {
		order.PackageStatus = row.PackageStatus
	}

	if err := validateOrder(&order); err != nil {
		return nil, fmt.Errorf("gecersiz siparis verisi: %w", err)
	}

	return &order, nil
}

func validateOrder(o *Order) error {
	if o.OrderID == "" {
		return fmt.Errorf("order_id bos olamaz")
	}
	if o.OrderNumber == "" {
		return fmt.Errorf("order_number bos olamaz")
	}
	if len(o.Lines) == 0 {
		return fmt.Errorf("siparis en az bir urun icermelidir")
	}
	return nil
}
