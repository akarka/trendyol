# .env dosyasındaki değişkenleri oku
$envFile = ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        $key, $value = $_.Split('=', 2)
        if ($key -and $value) {
            Set-Content "env:\$key" $value.Trim()
        }
    }
} else {
    Write-Warning ".env dosyasi bulunamadi. WEBHOOK_USERNAME ve WEBHOOK_PASSWORD manuel olarak ayarlanmali."
}

$json = '{
    "id": "curl-test-001",
    "orderNumber": "123456789",
    "packageStatus": "Created",
    "orderDate": "2026-03-01T20:45:00Z",
    "cargoProviderName": "Trendyol Express",
    "shipmentAddress": {
        "firstName": "Ahmet",
        "lastName": "Yilmaz",
        "address1": "Ataturk Mah. No:1",
        "city": "Ankara",
        "district": "Cankaya",
        "postalCode": "06000"
    },
    "lines": [
        {
            "productName": "Kablosuz Mouse",
            "barcode": "868123456789",
            "quantity": 1,
            "amount": 450.00
        }
    ]
}'

# Basic Auth için kimlik bilgilerini Base64'e çevir
$username = $env:WEBHOOK_USERNAME
$password = $env:WEBHOOK_PASSWORD
$encodedAuth = [System.Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes("${username}:${password}"))
$headers = @{
    "Authorization" = "Basic $encodedAuth"
}

$uri = "https://ruyxkthhvhbygwhwswri.supabase.co/functions/v1/trendyol-webhook"

Write-Host "Siparis gonderiliyor..." -ForegroundColor Cyan
Invoke-RestMethod -Uri $uri -Method Post -ContentType "application/json; charset=utf-8" -Body $json -Headers $headers
Write-Host "`nTamamlandi!" -ForegroundColor Green
