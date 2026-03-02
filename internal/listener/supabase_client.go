package listener

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/akarka/trendyol/config"
	"github.com/gorilla/websocket"
)

const (
	reconnectDelay    = 5 * time.Second
	maxReconnectDelay = 60 * time.Second
	heartbeatInterval = 30 * time.Second
	writeWait         = 10 * time.Second
)

type PhoenixMessage struct {
	Topic   string          `json:"topic"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
	Ref     string          `json:"ref"`
}

type PostgresChangesPayload struct {
	Type   string          `json:"type"`
	Record json.RawMessage `json:"record"`
}

func StartRealtimeSubscription(cfg *config.Config, dbChannel chan<- string) {
	delay := reconnectDelay

	for {
		err := subscribe(cfg, dbChannel)
		if err != nil {
			log.Printf("[REALTIME] Bağlantı/Abonelik hatası: %v. %v sonra yeniden denenecek.", err, delay)
			time.Sleep(delay)

			delay *= 2
			if delay > maxReconnectDelay {
				delay = maxReconnectDelay
			}
		} else {
			delay = reconnectDelay
		}
	}
}

func subscribe(cfg *config.Config, dbChannel chan<- string) error {
	host := strings.TrimPrefix(cfg.SupabaseURL, "https://")
	wsURL := url.URL{Scheme: "wss", Host: host, Path: "/realtime/v1/websocket", RawQuery: "apikey=" + cfg.SupabaseAnonKey + "&vsn=1.0.0"}
	
	log.Printf("[REALTIME] Supabase'e bağlanılıyor: %s", wsURL.String())

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket dial hatası: %w", err)
	}
	defer conn.Close()
	log.Println("[REALTIME] Bağlantı başarılı.")

	var refCounter uint64

	ref := fmt.Sprintf("%d", atomic.AddUint64(&refCounter, 1))
	subscriptionMsg := map[string]interface{}{
		"topic": "realtime:public:trendyol_orders",
		"event": "phx_join",
		"payload": map[string]interface{}{
			"config": map[string]interface{}{
				"postgres_changes": []map[string]interface{}{
					{
						"event":  "INSERT",
						"schema": "public",
						"table":  "trendyol_orders",
					},
				},
			},
		},
		"ref": ref,
	}

	if err := conn.WriteJSON(subscriptionMsg); err != nil {
		return fmt.Errorf("abonelik mesajı gönderilemedi: %w", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[REALTIME] Okuma hatası: %v", err)
				return
			}

			var msg PhoenixMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("[REALTIME] JSON parse hatası (dış): %v", err)
				continue
			}

			if msg.Event == "postgres_changes" {
				var payloadData map[string]interface{}
				if err := json.Unmarshal(msg.Payload, &payloadData); err != nil {
					log.Printf("[REALTIME] Payload parse hatası: %v", err)
					continue
				}

				// The payload structure can vary, but typically it contains a 'data' object
				// which has the actual change details.
				if dataMap, ok := payloadData["data"].(map[string]interface{}); ok {
					if typeVal, typeOk := dataMap["type"].(string); typeOk && typeVal == "INSERT" {
						if record, recordOk := dataMap["record"].(map[string]interface{}); recordOk {
							recordJSON, _ := json.Marshal(record)
							log.Println("[REALTIME] Yeni sipariş alındı, kanala gönderiliyor.")
							dbChannel <- string(recordJSON)
						}
					}
				} else if typeVal, ok := payloadData["type"].(string); ok && typeVal == "INSERT" {
					// Alternative payload structure
					if record, recordOk := payloadData["record"].(map[string]interface{}); recordOk {
						recordJSON, _ := json.Marshal(record)
						log.Println("[REALTIME] Yeni sipariş alındı, kanala gönderiliyor.")
						dbChannel <- string(recordJSON)
					}
				}
			}
		}
	}()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			ref := fmt.Sprintf("%d", atomic.AddUint64(&refCounter, 1))
			heartbeatMsg := map[string]interface{}{
				"topic":   "phoenix",
				"event":   "heartbeat",
				"payload": map[string]string{},
				"ref":     ref,
			}
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteJSON(heartbeatMsg); err != nil {
				return fmt.Errorf("heartbeat gönderilemedi: %w", err)
			}
		}
	}
}
