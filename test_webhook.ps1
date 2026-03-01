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

Write-Host "Siparis gonderiliyor..." -ForegroundColor Cyan
Invoke-RestMethod -Uri "http://localhost:8080/webhook" -Method Post -ContentType "application/json; charset=utf-8" -Body $json
Write-Host "`nTamamlandi!" -ForegroundColor Green
