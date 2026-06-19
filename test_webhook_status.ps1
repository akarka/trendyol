# Parametre: Statü seçimi (Varsayılan: Created)
param (
    [string]$Status = "Created"
)

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

$validStatuses = @("Created", "Cancelled", "Delivered", "UnSupplied")
if ($validStatuses -notcontains $Status) {
    Write-Error "Gecersiz statu! Lutfen sunlardan birini secin: $($validStatuses -join ', ')"
    exit 1
}

# Her çalıştırmada benzersiz bir ID oluştur
$randomNumber = Get-Random -Minimum 1000 -Maximum 9999
$uniqueId = "curl-test-$randomNumber-$(Get-Date -Format 'HHmmss')"

# Sipariş tarihini bugüne ayarla
$currentDate = Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ'

# Payload'u hazırla
$json = @"
{
    "id": "$uniqueId",
    "orderNumber": "987654321",
    "packageStatus": "$Status",
    "orderDate": "$currentDate",
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

if (-not $username -or -not $password) {
    Write-Error "Webhook kimlik bilgileri eksik. Lutfen .env dosyasini kontrol edin."
    exit 1
}

$encodedAuth = [System.Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes("${username}:${password}"))
$headers = @{
    "Authorization" = "Basic $encodedAuth"
}

$uri = "http://localhost:8080/webhook/trendyol"

Write-Host "Benzersiz '$Status' statulu siparis gonderiliyor (ID: $uniqueId)..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri $uri -Method Post -ContentType "application/json; charset=utf-8" -Body $json -Headers $headers
    Write-Host "Sunucu Yaniti: " -NoNewline
    Write-Host $response -ForegroundColor Yellow
    Write-Host "`nTamamlandi!" -ForegroundColor Green
} catch {
    Write-Error "İstek sirasinda hata olustu: $_"
}
