// internal/listener/supabase_client.go
package listener
 
import (
    "log"
    "github.com/akarka/trendyol-print-relay/config"
)
 
func StartRealtimeSubscription(cfg *config.Config, dbChannel chan<- string) {
    log.Println("[LISTENER] Supabase Realtime is currently disabled as per request.")
    // Mock listener that does nothing
    select {} 
}
