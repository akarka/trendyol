// internal/printer/escpos_printer.go
package printer

import (
	"fmt"
	"os"

	"github.com/akarka/trendyol-print-relay/internal/parser"
	"github.com/kenshin54/go-escpos"
)

// Print prints the order to the ESC/POS printer
func Print(devicePath string, order *parser.Order) error {
	f, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("printer device açılırken hata: %w", err)
	}
	defer f.Close()

	p := escpos.New(f)
	p.Init()

	// Header
	p.SetEmphasis(1)
	p.Write("TRENDYOL SIPARIS")
	p.SetEmphasis(0)
	p.Formfeed()

	// Order Details
	p.Write(fmt.Sprintf("Siparis No: %s", order.OrderNumber))
	p.Formfeed()
	p.Write(fmt.Sprintf("Tarih: %s", order.CreatedAt.Format("02.01.2006 15:04")))
	p.Formfeed()

	p.Write("--------------------------------")
	p.Formfeed()

	// Items
	for _, line := range order.Lines {
		p.Write(fmt.Sprintf("%dx %s", line.Quantity, line.ProductName))
		p.Formfeed()
		p.Write(fmt.Sprintf("Barkod: %s", line.Barcode))
		p.Formfeed()
	}

	p.Write("--------------------------------")
	p.Formfeed()

	// Shipment
	p.Write("Teslimat Adresi:")
	p.Formfeed()
	p.Write(fmt.Sprintf("%s %s", order.ShipmentInfo.FirstName, order.ShipmentInfo.LastName))
	p.Formfeed()
	p.Write(order.ShipmentInfo.Address1)
	p.Formfeed()
	p.Write(fmt.Sprintf("%s / %s", order.ShipmentInfo.District, order.ShipmentInfo.City))
	p.Formfeed()

	p.Formfeed()
	p.Formfeed()
	p.Cut()

	return nil
}
