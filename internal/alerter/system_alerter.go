// internal/alerter/system_alerter.go
package alerter

import (
	"fmt"
	"log"
)

// NotifySuccess logs success and triggers a notification
func NotifySuccess(orderNumber string) {
	log.Printf("[SUCCESS] Siparis basariyla yazdirildi: %s", orderNumber)
	// Basic OS notification via STDOUT/BEEP
	fmt.Print("\a") // System beep
}

// NotifyError logs error and triggers a notification
func NotifyError(orderNumber string) {
	log.Printf("[ERROR] Siparis yazdirma hatasi: %s", orderNumber)
	// Alert sound (repeated)
	fmt.Print("\a\a\a")
}
