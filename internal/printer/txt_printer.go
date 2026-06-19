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

	// Sipariş durumunu Türkçeye çevir
	statusTr := order.PackageStatus
	switch order.PackageStatus {
	case "Created":
		statusTr = "Yeni Sipariş"
	case "Cancelled":
		statusTr = "İptal Edildi"
	case "Delivered":
		statusTr = "Teslim Edildi"
	case "UnSupplied":
		statusTr = "Tedarik Edilemedi"
	}

	builder.WriteString("========================================\n")
	builder.WriteString(fmt.Sprintf("Tarih      : %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Sipariş No : %s\n", order.OrderNumber))
	builder.WriteString(fmt.Sprintf("Durum      : %s\n", statusTr))
	
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
		return fmt.Errorf("çıktı dosyası açılamadı: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(builder.String()); err != nil {
		return fmt.Errorf("çıktı dosyasına yazılamadı: %w", err)
	}

	fmt.Printf("[TXT YAZICI] %s numaralı sipariş %s dosyasına eklendi\n", order.OrderNumber, filePath)
	return nil
}
