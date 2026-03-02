package printer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/akarka/trendyol/internal/parser"
)

// PrintToTextFile formats the order details and appends them to a specified text file.
func PrintToTextFile(filePath string, order *parser.Order) error {
	var builder strings.Builder

	builder.WriteString("----------------------------------------\n")
	builder.WriteString(fmt.Sprintf("Sipariş Tarihi: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Sipariş No: %s\n", order.OrderNumber))
	builder.WriteString("----------------------------------------\n\n")

	for _, line := range order.Lines {
		lineStr := fmt.Sprintf("%d x %s\n", line.Quantity, line.ProductName)
		builder.WriteString(lineStr)
		// Example for price, you can add more details
		// priceStr := fmt.Sprintf("   Birim Fiyat: %.2f TL\n", line.Price)
		// builder.WriteString(priceStr)
	}

	builder.WriteString("\n----------------------------------------\n\n")

	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("metin dosyası açılamadı: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(builder.String()); err != nil {
		return fmt.Errorf("metin dosyasına yazılamadı: %w", err)
	}

	fmt.Printf("[DIGITAL PRINTER] Order %s written to %s\n", order.OrderNumber, filePath)
	return nil
}
