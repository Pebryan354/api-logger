package apilogger_test

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	apilogger "github.com/Pebryan354/api-logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	return db
}

func TestLogger_DefaultMigrationAndLogging(t *testing.T) {
	db := setupTestDB(t)

	logger, err := apilogger.New(db)
	if err != nil {
		t.Fatalf("apilogger.New failed: %v", err)
	}

	// Verify tables were created
	if !db.Migrator().HasTable(apilogger.DefaultAPITableName) {
		t.Fatalf("expected table %s to exist", apilogger.DefaultAPITableName)
	}
	if !db.Migrator().HasTable(apilogger.DefaultCallbackTableName) {
		t.Fatalf("expected table %s to exist", apilogger.DefaultCallbackTableName)
	}

	// Test LogAPI
	ctx := context.Background()
	resendID := uint64(999)
	apiEntry := &apilogger.ResponseAPI{
		Type:              "CREATE_VA",
		Endpoint:          "https://api.payment.example.com/v1/va",
		Header:            `{"Authorization": "Bearer token123"}`,
		Param:             `{"bank": "BCA", "amount": 50000}`,
		HTTPMethod:        "POST",
		HTTPCode:          "200",
		Result:            `{"status": "success", "va_number": "1234567890"}`,
		DBTable:           "transactions",
		DBTableField:      "trx_code",
		DBTableFieldValue: "TRX-2026-001",
		DBTrans:           "commit",
		ResendID:          &resendID,
	}

	if err := logger.LogAPI(ctx, apiEntry); err != nil {
		t.Fatalf("LogAPI failed: %v", err)
	}

	if apiEntry.ID == 0 {
		t.Errorf("expected auto-increment ID to be set, got 0")
	}
	if apiEntry.Curl == "" {
		t.Errorf("expected curl to be auto-generated")
	}
	if apiEntry.ScriptPath == "" || apiEntry.ScriptLine == "" {
		t.Errorf("expected caller script info to be captured automatically, got path=%s line=%s", apiEntry.ScriptPath, apiEntry.ScriptLine)
	}

	// Test LogCallback
	refID := uint64(1001)
	callbackEntry := &apilogger.ResponseCallback{
		Type:       "VA_PAYMENT",
		Header:     `{"Content-Type": "application/json"}`,
		Param:      `{"va_number": "1234567890", "paid_amount": 50000}`,
		HTTPMethod: "POST",
		HTTPCode:   "200",
		Result:     `{"status": "ACK"}`,
		RefType:    apilogger.RefTypePayment,
		RefID:      &refID,
	}

	if err := logger.LogCallback(ctx, callbackEntry); err != nil {
		t.Fatalf("LogCallback failed: %v", err)
	}

	if callbackEntry.ID == 0 {
		t.Errorf("expected callback auto-increment ID to be set, got 0")
	}
}

func TestLogger_CustomTableNames(t *testing.T) {
	db := setupTestDB(t)

	customAPI := "custom_api_logs"
	customCallback := "custom_callback_logs"

	logger, err := apilogger.New(db,
		apilogger.WithAPITableName(customAPI),
		apilogger.WithCallbackTableName(customCallback),
	)
	if err != nil {
		t.Fatalf("apilogger.New failed: %v", err)
	}

	if !db.Migrator().HasTable(customAPI) {
		t.Errorf("expected custom table %s to exist", customAPI)
	}
	if !db.Migrator().HasTable(customCallback) {
		t.Errorf("expected custom table %s to exist", customCallback)
	}

	apiEntry := &apilogger.ResponseAPI{
		Type:       "INQUIRY",
		HTTPMethod: "GET",
		Endpoint:   "https://api.example.com/inquiry",
		HTTPCode:   "200",
	}
	if err := logger.LogAPI(context.Background(), apiEntry); err != nil {
		t.Fatalf("LogAPI with custom table failed: %v", err)
	}

	var count int64
	db.Table(customAPI).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row in %s, got %d", customAPI, count)
	}
}

func TestLogger_Async(t *testing.T) {
	db := setupTestDB(t)

	var wg sync.WaitGroup
	wg.Add(1)

	logger, err := apilogger.New(db,
		apilogger.WithAsync(true),
		apilogger.WithErrorHandler(func(err error) {
			t.Errorf("unexpected async error: %v", err)
		}),
	)
	if err != nil {
		t.Fatalf("apilogger.New failed: %v", err)
	}

	apiEntry := &apilogger.ResponseAPI{
		Type:       "ASYNC_TEST",
		HTTPMethod: "POST",
		Endpoint:   "https://api.example.com",
	}

	if err := logger.LogAPI(context.Background(), apiEntry); err != nil {
		t.Fatalf("LogAPI async failed: %v", err)
	}

	// Allow goroutine to execute
	time.Sleep(100 * time.Millisecond)

	var count int64
	db.Table(apilogger.DefaultAPITableName).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row inserted asynchronously, got %d", count)
	}
}

func TestCurlHelpers(t *testing.T) {
	req, err := http.NewRequest("POST", "https://api.example.com/pay", bytes.NewBufferString(`{"amount":100}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer xyz")
	req.Header.Set("Content-Type", "application/json")

	curlCmd := apilogger.BuildCurlFromRequest(req)
	expectedContains := []string{
		"curl -X POST 'https://api.example.com/pay'",
		"-H 'Authorization: Bearer xyz'",
		"-H 'Content-Type: application/json'",
		"-d '{\"amount\":100}'",
	}

	for _, str := range expectedContains {
		if !bytes.Contains([]byte(curlCmd), []byte(str)) {
			t.Errorf("curl output missing expected part %q, got: %s", str, curlCmd)
		}
	}

	// Verify request body can still be read
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	if buf.String() != `{"amount":100}` {
		t.Errorf("request body was consumed and not restored, got: %s", buf.String())
	}
}

func TestToJSON(t *testing.T) {
	data := map[string]any{"status": "ok", "code": 200}
	str := apilogger.ToJSON(data)
	if str != `{"code":200,"status":"ok"}` {
		t.Errorf("unexpected json result: %s", str)
	}

	if apilogger.ToJSON(nil) != "" {
		t.Errorf("expected empty string for nil")
	}

	if apilogger.ToJSON("already string") != "already string" {
		t.Errorf("expected string as-is")
	}
}
