package apilogger

import (
	"time"
)

const (
	// DefaultAPITableName is the default table name for ResponseAPI to prevent conflicts with main app tables.
	DefaultAPITableName = "log_response_api"

	// DefaultCallbackTableName is the default table name for ResponseCallback to prevent conflicts with main app tables.
	DefaultCallbackTableName = "log_response_callback"
)

// ResponseAPI records outbound/external API requests, responses, and tracing context.
type ResponseAPI struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Type              string    `gorm:"size:100;not null;index" json:"type"`
	Curl              string    `gorm:"type:text" json:"curl,omitempty"`
	Endpoint          string    `gorm:"type:text" json:"endpoint,omitempty"`
	Header            string    `gorm:"type:text" json:"header,omitempty"`
	Param             string    `gorm:"type:text" json:"param,omitempty"`
	HTTPMethod        string    `gorm:"size:20" json:"http_method,omitempty"`
	HTTPCode          string    `gorm:"size:10" json:"http_code,omitempty"`
	Result            string    `gorm:"type:text" json:"result,omitempty"`
	DBTable           string    `gorm:"size:100" json:"db_table,omitempty"`
	DBTableField      string    `gorm:"size:100" json:"db_table_field,omitempty"`
	DBTableFieldValue string    `gorm:"size:100" json:"db_table_field_value,omitempty"`
	DBTrans           string    `gorm:"size:250" json:"db_trans,omitempty"`
	ScriptPath        string    `gorm:"type:text" json:"script_path,omitempty"`
	ScriptFetchClass  string    `gorm:"type:text" json:"script_fetch_class,omitempty"`
	ScriptFetchMethod string    `gorm:"type:text" json:"script_fetch_method,omitempty"`
	ScriptLine        string    `gorm:"type:text" json:"script_line,omitempty"`
	ResendID          *uint64   `gorm:"index" json:"resend_id,omitempty"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns default table name for ResponseAPI
func (ResponseAPI) TableName() string {
	return DefaultAPITableName
}

// Common reference types for callback events.
const (
	RefTypePayment      = "payment"
	RefTypeDisbursement = "disbursement"
	RefTypeQRIS         = "qris"
	RefTypeTopup        = "topup"
)

// ResponseCallback records inbound webhook/callback notifications and results.
type ResponseCallback struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Type       string    `gorm:"size:100;not null;index" json:"type"`
	Header     string    `gorm:"type:text;not null" json:"header"`
	Param      string    `gorm:"type:text" json:"param,omitempty"`
	HTTPMethod string    `gorm:"size:20;not null" json:"http_method"`
	HTTPCode   string    `gorm:"size:10;not null" json:"http_code"`
	Result     string    `gorm:"type:text" json:"result,omitempty"`
	RefType    string    `gorm:"size:50;index" json:"ref_type,omitempty"`
	RefID      *uint64   `gorm:"index" json:"ref_id,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns default table name for ResponseCallback
func (ResponseCallback) TableName() string {
	return DefaultCallbackTableName
}
