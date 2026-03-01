-- supabase/migrations/20260301_init.sql

CREATE TABLE trendyol_orders (
    uuid            UUID                     PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        VARCHAR(255)             NOT NULL,
    order_number    VARCHAR(255)             NOT NULL,
    package_status  VARCHAR(50)              NOT NULL,
    payload         JSONB                    NOT NULL,
    created_at      TIMESTAMPTZ              DEFAULT NOW(),
    updated_at      TIMESTAMPTZ              DEFAULT NOW(),
 
    -- Idempotency: aynı paket + statü çifti yalnızca bir kez saklanır
    CONSTRAINT unique_order_status UNIQUE (order_id, package_status)
);
 
-- updated_at otomatik güncelleme
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
 
CREATE TRIGGER trendyol_orders_updated_at
    BEFORE UPDATE ON trendyol_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Sık sorgulanan alanlar için performans indeksleri
CREATE INDEX idx_trendyol_orders_order_id       ON trendyol_orders (order_id);
CREATE INDEX idx_trendyol_orders_created_at     ON trendyol_orders (created_at DESC);
CREATE INDEX idx_trendyol_orders_status         ON trendyol_orders (package_status);
 
-- Payload içindeki yüksek kardinaliteli alanlar için (opsiyonel, trafik hacmine göre)
CREATE INDEX idx_trendyol_payload_merchant
    ON trendyol_orders ((payload->>'merchantId'));

-- RLS aktif et
ALTER TABLE trendyol_orders ENABLE ROW LEVEL SECURITY;
 
-- Yalnızca Edge Function (service_role) INSERT yapabilir
CREATE POLICY "edge_function_insert"
    ON trendyol_orders FOR INSERT
    TO service_role
    WITH CHECK (true);
 
-- Anon key yalnızca SELECT yapabilir (Go listener Realtime için)
CREATE POLICY "listener_select"
    ON trendyol_orders FOR SELECT
    TO anon
    USING (true);
