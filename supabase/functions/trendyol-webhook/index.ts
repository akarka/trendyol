import { serve } from "https://deno.land/std@0.168.0/http/server.ts"
import { createClient } from "https://esm.sh/@supabase/supabase-js@2"

const SUPABASE_URL = Deno.env.get("SUPABASE_URL")!
const SUPABASE_SERVICE_KEY = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!
const WEBHOOK_USERNAME = Deno.env.get("WEBHOOK_USERNAME")!
const WEBHOOK_PASSWORD = Deno.env.get("WEBHOOK_PASSWORD")!

const PG_UNIQUE_VIOLATION = "23505"

function isValidBasicAuth(header: string | null): boolean {
  if (!header?.startsWith("Basic ")) return false
  try {
    const decoded = atob(header.slice(6))
    const [user, pass] = decoded.split(":")
    return user === WEBHOOK_USERNAME && pass === WEBHOOK_PASSWORD
  } catch {
    return false
  }
}

interface TrendyolPayload {
  id: string
  orderNumber: string
  packageStatus: string
  [key: string]: unknown
}

function validatePayload(data: unknown): data is TrendyolPayload {
  if (typeof data !== "object" || data === null) return false
  const d = data as Record<string, unknown>
  return (
    typeof d.id === "string" && d.id.length > 0 &&
    typeof d.orderNumber === "string" && d.orderNumber.length > 0 &&
    typeof d.packageStatus === "string" && d.packageStatus.length > 0
  )
}

serve(async (req: Request) => {
  if (req.method !== "POST") {
    return new Response("Method Not Allowed", { status: 405 })
  }

  if (!isValidBasicAuth(req.headers.get("Authorization"))) {
    return new Response("Unauthorized", { status: 401 })
  }

  let payload: unknown
  try {
    payload = await req.json()
  } catch {
    return new Response("Bad Request: invalid JSON", { status: 400 })
  }

  if (!validatePayload(payload)) {
    return new Response(
      "Bad Request: id, orderNumber ve packageStatus alanları zorunludur",
      { status: 400 }
    )
  }

  const supabase = createClient(SUPABASE_URL, SUPABASE_SERVICE_KEY)

  const { error } = await supabase
    .from("trendyol_orders")
    .insert({
      order_id: payload.id,
      order_number: payload.orderNumber,
      package_status: payload.packageStatus,
      payload: payload,
    })

  if (error) {
    if (error.code === PG_UNIQUE_VIOLATION) {
      console.info(`[DUPLICATE] order_id=${payload.id} status=${payload.packageStatus}`)
      // For duplicates, it's okay to return 200 as it's expected behavior
      return new Response("OK (Duplicate)", { status: 200 })
    } else {
      console.error(`[DB_ERROR] code=${error.code} message=${error.message}`, payload)
      // For debugging, return the actual error message
      return new Response(
        `Database Error: ${error.message}`,
        { status: 500, headers: { "Content-Type": "text/plain" } }
      )
    }
  }

  // On successful insert
  return new Response("OK (Inserted)", { status: 200 })
})
