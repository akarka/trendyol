package alerter

import (
	"log"
)

func NotifySuccess(orderNumber string) {
	// TODO: Trigger a basic OS sound or standard output notification
	log.Printf("[BİLDİRİM] Başarılı: %s numaralı sipariş yazdırıldı.\n", orderNumber)
}

func NotifyError(orderNumber string) {
	// TODO: Trigger a basic OS sound or standard output notification
	log.Printf("[BİLDİRİM] Hata: %s numaralı sipariş yazdırılamadı.\n", orderNumber)
}
