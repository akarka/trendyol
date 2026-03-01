// internal/printer/digital_printer.go
package printer

import (
	"fmt"
	"os"
	"strings"

	"github.com/akarka/trendyol-print-relay/internal/parser"
)

// PrintToText generates a human-readable digital representation of the order
func PrintToText(filePath string, order *parser.Order) error {
	var sb strings.Builder

	sb.WriteString("==========================================
")
	sb.WriteString("            TRENDYOL SIPARIS             
")
	sb.WriteString("==========================================
")
	sb.WriteString(fmt.Sprintf("Siparis No:    %s
", order.OrderNumber))
	sb.WriteString(fmt.Sprintf("Tarih:         %s
", order.CreatedAt.Format("02.01.2006 15:04")))
	sb.WriteString(fmt.Sprintf("Paket Durumu:  %s
", order.PackageStatus))
	sb.WriteString(fmt.Sprintf("Kargo Firmasi: %s
", order.CargoProvider))
	sb.WriteString("------------------------------------------
")
	sb.WriteString(fmt.Sprintf("%-25s | %-3s | %-7s
", "Urun Adi", "Adet", "Fiyat"))
	sb.WriteString("------------------------------------------
")

	for _, line := range order.Lines {
		name := line.ProductName
		if len(name) > 24 {
			name = name[:21] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-25s | %-4d | %-7.2f
", name, line.Quantity, line.Price))
		sb.WriteString(fmt.Sprintf("  Barkod: %s
", line.Barcode))
	}

	sb.WriteString("------------------------------------------
")
	sb.WriteString("TESLIMAT ADRESI:
")
	sb.WriteString(fmt.Sprintf("%s %s
", order.ShipmentInfo.FirstName, order.ShipmentInfo.LastName))
	sb.WriteString(fmt.Sprintf("%s
", order.ShipmentInfo.Address1))
	sb.WriteString(fmt.Sprintf("%s / %s
", order.ShipmentInfo.District, order.ShipmentInfo.City))
	sb.WriteString(fmt.Sprintf("Posta Kodu: %s
", order.ShipmentInfo.PostalCode))
	sb.WriteString("==========================================
")

	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}
