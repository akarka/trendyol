-- Enable Realtime for the trendyol_orders table so the Go listener can receive notifications.
ALTER PUBLICATION supabase_realtime ADD TABLE trendyol_orders;
