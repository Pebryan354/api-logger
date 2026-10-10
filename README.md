# API Logger

Library untuk mencatat riwayat pemanggilan API eksternal (`ResponseAPI`) dan penerimaan callback/webhook sistem (`ResponseCallback`) ke database menggunakan [GORM](https://gorm.io).

Library ini dirancang fleksibel untuk berbagai database yang didukung GORM (**MariaDB**, **MySQL**, **PostgreSQL**, **SQLite**, dll), serta dilengkapi dengan **AutoMigrate** otomatis dan nama tabel bawaan yang aman agar tidak bentrok dengan tabel aplikasi utama.

---

## Fitur Utama

- **Otomatis AutoMigrate**: Tabel dibuat otomatis saat library diinisialisasi (bisa dinonaktifkan jika diinginkan).
- **Aman dari Konflik Nama Tabel**: Menggunakan default nama tabel `log_response_api` dan `log_response_callback`, serta mendukung kustomisasi nama tabel.
- **Auto Capture Caller**: Otomatis mendeteksi file path (`script_path`), nama class/struct (`script_fetch_class`), fungsi/method (`script_fetch_method`), dan baris kode (`script_line`) pemanggil tanpa perlu diisi manual.
- **Auto cURL Generator**: Menghasilkan perintah `curl` lengkap yang siap di-copy langsung ke terminal untuk debugging.
- **Mendukung Mode Sinkron & Asinkron**: Dapat berjalan di background goroutine agar tidak membebani latency respons user.
- **Database Agnostik**: Bekerja mulus di MariaDB, MySQL, PostgreSQL, maupun SQLite.

---

## Instalasi

```bash
go get github.com/Pebryan354/api-logger
```

---

## Cara Penggunaan

### 1. Inisialisasi

```go
package main

import (
	"log"

	apilogger "github.com/Pebryan354/api-logger"
	"gorm.io/driver/mysql" // atau driver database lain (postgres, sqlite)
	"gorm.io/gorm"
)

func main() {
	// Contoh koneksi MariaDB / MySQL
	dsn := "user:password@tcp(127.0.0.1:3306)/my_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}

	// Inisialisasi logger (otomatis melakukan AutoMigrate)
	logger, err := apilogger.New(db)
	if err != nil {
		log.Fatalf("gagal inisialisasi api-logger: %v", err)
	}

	_ = logger
}
```

#### Opsi Konfigurasi Tambahan (Opsional)

```go
logger, err := apilogger.New(db,
	// Ubah nama tabel jika ingin kustom
	apilogger.WithAPITableName("custom_response_api"),
	apilogger.WithCallbackTableName("custom_response_callback"),

	// Nonaktifkan automigrate jika migrasi diatur terpisah
	apilogger.WithAutoMigrate(true),

	// Aktifkan mode asinkron (background goroutine)
	apilogger.WithAsync(true),
	apilogger.WithErrorHandler(func(err error) {
		log.Printf("logger error: %v", err)
	}),
)
```

---

### 2. Mencatat Log Response API (Outgoing Request)

Gunakan `LogAPI` ketika aplikasi Anda memanggil layanan pihak ketiga (contoh: request Create Virtual Account):

```go
ctx := context.Background()

err := logger.LogAPI(ctx, &apilogger.ResponseAPI{
	Type:              "CREATE_VA",
	Endpoint:          "https://api.paymentgateway.com/v1/va/create",
	Header:            apilogger.ToJSON(map[string]string{"Authorization": "Bearer xxx"}),
	Param:             apilogger.ToJSON(map[string]any{"bank": "BCA", "amount": 100000}),
	HTTPMethod:        "POST",
	HTTPCode:          "200",
	Result:            `{"status":"SUCCESS","va_number":"880812345678"}`,
	DBTable:           "transactions",
	DBTableField:      "trx_code",
	DBTableFieldValue: "TRX-2026-001",
	DBTrans:           "commit",
	// script_path, script_fetch_class, script_fetch_method, script_line, dan curl
	// akan terisi OTOMATIS jika dikosongkan!
})
if err != nil {
	log.Printf("gagal simpan log api: %v", err)
}
```

Jika ini adalah percobaan pengiriman ulang (resend), Anda dapat menyertakan `ResendID`:

```go
prevLogID := uint64(123)
entry := &apilogger.ResponseAPI{
	Type:     "CREATE_VA",
	ResendID: &prevLogID,
	// ...
}
```

---

### 3. Mencatat Log Response Callback (Inbound Webhook)

Gunakan `LogCallback` saat menerima webhook / notifikasi dari pihak ketiga:

```go
refID := uint64(5001) // Contoh ID referensi transaksi/withdraw di database Anda

err := logger.LogCallback(ctx, &apilogger.ResponseCallback{
	Type:       "VA_PAYMENT",
	Header:     `{"Content-Type": "application/json"}`,
	Param:      `{"va_number": "880812345678", "amount": 100000, "status": "PAID"}`,
	HTTPMethod: "POST",
	HTTPCode:   "200",
	Result:     `{"status": "ACK"}`,
	RefType:    apilogger.RefTypePayment, // atau "disbursement", "qris", "topup", atau tipe kustom
	RefID:      &refID,
})
if err != nil {
	log.Printf("gagal simpan log callback: %v", err)
}
```

---

### 4. Helper Tambahan

#### Generate cURL Langsung dari `*http.Request`

```go
req, _ := http.NewRequest("POST", "https://api.example.com", body)
curlString := apilogger.BuildCurlFromRequest(req)
// Output: curl -X POST 'https://api.example.com' -H '...' -d '...'
```

#### Helper `ToJSON`

Mengubah struct, slice, atau map menjadi string JSON secara aman:

```go
jsonStr := apilogger.ToJSON(myPayload)
```
