// internal/printer/digital_printer.go
package printer

import (
	"fmt"
	"os"
	"strings"

	"github.com/akarka/trendyol/internal/parser"
)

// PrintToText generates a human-readable digital representation of the order
func PrintToText(filePath string, order *parser.Order) error {
	var sb strings.Builder

	sb.WriteString("==========================================\n")
	sb.WriteString("            TRENDYOL SIPARIS             \n")
	sb.WriteString("==========================================\n")
	sb.WriteString(fmt.Sprintf("Siparis No:    %s\n", order.OrderNumber))
	sb.WriteString(fmt.Sprintf("Tarih:         %s\n", order.CreatedAt.Format("02.01.2006 15:04")))
	sb.WriteString(fmt.Sprintf("Paket Durumu:  %s\n", order.PackageStatus))
	sb.WriteString(fmt.Sprintf("Kargo Firmasi: %s\n", order.CargoProvider))
	sb.WriteString("------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("%-25s | %-3s | %-7s\n", "Urun Adi", "Adet", "Fiyat"))
	sb.WriteString("------------------------------------------\n")

	for _, line := range order.Lines {
		name := line.ProductName
		if len(name) > 24 {
			name = name[:21] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-25s | %-4d | %-7.2f\n", name, line.Quantity, line.Price))
		sb.WriteString(fmt.Sprintf("  Barkod: %s\n", line.Barcode))
	}

	sb.WriteString("------------------------------------------\n")
	sb.WriteString("TESLIMAT ADRESI:\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", order.ShipmentInfo.FirstName, order.ShipmentInfo.LastName))
	sb.WriteString(fmt.Sprintf("%s\n", order.ShipmentInfo.Address1))
	sb.WriteString(fmt.Sprintf("%s / %s\n", order.ShipmentInfo.District, order.ShipmentInfo.City))
	sb.WriteString(fmt.Sprintf("Posta Kodu: %s\n", order.ShipmentInfo.PostalCode))
	sb.WriteString("==========================================\n")

	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}
