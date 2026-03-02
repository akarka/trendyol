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

func ParseOrder(raw string) (*Order, error) {
	var order Order
	if err := json.Unmarshal([]byte(raw), &order); err != nil {
		return nil, fmt.Errorf("JSON parse hatasi: %w", err)
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
