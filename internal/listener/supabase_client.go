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
	Topic   string      `json:"topic"`
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
	Ref     string      `json:"ref"`
}

type PostgresChangesPayload struct {
	Type   string                 `json:"type"`
	Schema string                 `json:"schema"`
	Table  string                 `json:"table"`
	Record map[string]interface{} `json:"record"`
}

func StartRealtimeSubscription(cfg *config.Config, dbChannel chan<- string) {
	delay := reconnectDelay

	for {
		err := subscribe(cfg, dbChannel)
		if err != nil {
			log.Printf("[REALTIME] Bağlantı hatası: %v. %v sonra yeniden denenecek.", err, delay)
			time.Sleep(delay)

			delay *= 2
			if delay > maxReconnectDelay {
				delay = maxReconnectDelay
			}
		} else {
			delay = reconnectDelay // Başarılı bağlantıda sıfırla
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

	// Subscribe to changes
	ref := fmt.Sprintf("%d", atomic.AddUint64(&refCounter, 1))
	subscriptionMsg := PhoenixMessage{
		Topic: "realtime:public:trendyol_orders",
		Event: "phx_join",
		Payload: map[string]interface{}{
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
		Ref: ref,
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
			
			// Gelen INSERT eventlerini işle
			if msg.Event == "postgres_changes" {
				payloadData, err := json.Marshal(msg.Payload)
				if err != nil {
					log.Printf("[REALTIME] Payload marshal hatası: %v", err)
					continue
				}

				var changes []PostgresChangesPayload
				if err := json.Unmarshal(payloadData, &changes); err != nil {
						// Bazen tek bir nesne olarak gelebilir
						var singleChange PostgresChangesPayload
						if err2 := json.Unmarshal(payloadData, &singleChange); err2 == nil {
								changes = []PostgresChangesPayload{singleChange}
						} else {
								log.Printf("[REALTIME] Payload parse hatası (iç): %v", err)
								continue
						}
				}
				
				for _, change := range changes {
					if change.Type == "INSERT" {
						recordJSON, err := json.Marshal(change.Record)
						if err != nil {
							log.Printf("[REALTIME] Record marshal hatası: %v", err)
							continue
						}
						log.Println("[REALTIME] Yeni sipariş alındı, kanala gönderiliyor.")
						dbChannel <- string(recordJSON)
					}
				}
			}
		}
	}()

	// Heartbeat
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			ref := fmt.Sprintf("%d", atomic.AddUint64(&refCounter, 1))
			heartbeatMsg := PhoenixMessage{Topic: "phoenix", Event: "heartbeat", Payload: map[string]string{}, Ref: ref}
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteJSON(heartbeatMsg); err != nil {
				return fmt.Errorf("heartbeat gönderilemedi: %w", err)
			}
		}
	}
}
