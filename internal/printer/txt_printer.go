package printer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/akarka/trendyol/internal/parser"
)

// PrintToTXT formats the order details and appends them to a local text file.
func PrintToTXT(order *parser.Order) error {
	filePath := "output.txt"
	var builder strings.Builder

	builder.WriteString("========================================\n")
	builder.WriteString(fmt.Sprintf("Tarih      : %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Sipariş No : %s\n", order.OrderNumber))
	
	customerName := "Bilinmiyor"
	if order.ShipmentInfo.FirstName != "" || order.ShipmentInfo.LastName != "" {
		customerName = strings.TrimSpace(fmt.Sprintf("%s %s", order.ShipmentInfo.FirstName, order.ShipmentInfo.LastName))
	}
	builder.WriteString(fmt.Sprintf("Müşteri    : %s\n", customerName))
	builder.WriteString("----------------------------------------\n")
	builder.WriteString("Ürünler:\n")

	var totalAmount float64
	for _, line := range order.Lines {
		builder.WriteString(fmt.Sprintf("%d x %s (%.2f TL)\n", line.Quantity, line.ProductName, line.Price))
		totalAmount += line.Price * float64(line.Quantity)
	}

	builder.WriteString("----------------------------------------\n")
	builder.WriteString(fmt.Sprintf("Toplam     : %.2f TL\n", totalAmount))
	builder.WriteString("========================================\n\n")

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(builder.String()); err != nil {
		return fmt.Errorf("failed to write to output file: %w", err)
	}

	fmt.Printf("[TXT PRINTER] Order %s appended to %s\n", order.OrderNumber, filePath)
	return nil
}
