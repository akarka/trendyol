package alerter

import (
	"log"
)

func NotifySuccess(orderNumber string) {
	// TODO: Trigger a basic OS sound or standard output notification
	log.Printf("[ALERTER] Success: Order %s printed successfully.\n", orderNumber)
}

func NotifyError(orderNumber string) {
	// TODO: Trigger a basic OS sound or standard output notification
	log.Printf("[ALERTER] Error: Failed to print order %s.\n", orderNumber)
}
