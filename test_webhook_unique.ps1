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

# Her çalıştırmada benzersiz bir ID oluştur
$randomNumber = Get-Random -Minimum 1000 -Maximum 9999
$uniqueId = "curl-test-$randomNumber-$(Get-Date -Format 'HHmmss')"

$json = @"
{
    "id": "$uniqueId",
    "orderNumber": "987654321",
    "packageStatus": "Created",
    "orderDate": "2026-03-02T10:00:00Z",
    "cargoProviderName": "Trendyol Express",
    "shipmentAddress": {
        "firstName": "Ayşe",
        "lastName": "Demir",
        "address1": "Cumhuriyet Mah. No:10",
        "city": "İzmir",
        "district": "Bornova",
        "postalCode": "35000"
    },
    "lines": [
        {
            "productName": "Bluetooth Klavye",
            "barcode": "868987654321",
            "quantity": 1,
            "amount": 750.00
        },
        {
            "productName": "USB-C Şarj Kablosu",
            "barcode": "868112233445",
            "quantity": 2,
            "amount": 120.00
        }
    ]
}
"@

# Basic Auth için kimlik bilgilerini Base64'e çevir
$username = $env:WEBHOOK_USERNAME
$password = $env:WEBHOOK_PASSWORD
$encodedAuth = [System.Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes("${username}:${password}"))
$headers = @{
    "Authorization" = "Basic $encodedAuth"
}

$uri = "https://ruyxkthhvhbygwhwswri.supabase.co/functions/v1/trendyol-webhook"

Write-Host "Benzersiz siparis gonderiliyor (ID: $uniqueId)..." -ForegroundColor Cyan
Invoke-RestMethod -Uri $uri -Method Post -ContentType "application/json; charset=utf-8" -Body $json -Headers $headers
Write-Host "`nTamamlandi!" -ForegroundColor Green
