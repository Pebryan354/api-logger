package apilogger

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Logger is the main client for recording API and callback responses to database.
type Logger struct {
	db  *gorm.DB
	cfg Config
}

// New creates a new Logger instance.
// By default, AutoMigrate is enabled and creates tables using default names.
func New(db *gorm.DB, opts ...Option) (*Logger, error) {
	if db == nil {
		return nil, errors.New("gorm db instance cannot be nil")
	}

	cfg := Config{
		APITableName:      DefaultAPITableName,
		CallbackTableName: DefaultCallbackTableName,
		AutoMigrate:       true,
		CallerSkip:        3,
		OnError:           func(err error) {},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	l := &Logger{
		db:  db,
		cfg: cfg,
	}

	if cfg.AutoMigrate {
		if err := l.AutoMigrate(); err != nil {
			return nil, err
		}
	}

	return l, nil
}

// AutoMigrate creates or updates the database tables for ResponseAPI and ResponseCallback.
func (l *Logger) AutoMigrate() error {
	if l.cfg.APITableName != "" && l.cfg.APITableName != DefaultAPITableName {
		if err := l.db.Table(l.cfg.APITableName).AutoMigrate(&ResponseAPI{}); err != nil {
			return fmt.Errorf("apilogger: failed to migrate api table '%s': %w", l.cfg.APITableName, err)
		}
	} else {
		if err := l.db.AutoMigrate(&ResponseAPI{}); err != nil {
			return fmt.Errorf("apilogger: failed to migrate default api table: %w", err)
		}
	}

	if l.cfg.CallbackTableName != "" && l.cfg.CallbackTableName != DefaultCallbackTableName {
		if err := l.db.Table(l.cfg.CallbackTableName).AutoMigrate(&ResponseCallback{}); err != nil {
			return fmt.Errorf("apilogger: failed to migrate callback table '%s': %w", l.cfg.CallbackTableName, err)
		}
	} else {
		if err := l.db.AutoMigrate(&ResponseCallback{}); err != nil {
			return fmt.Errorf("apilogger: failed to migrate default callback table: %w", err)
		}
	}

	return nil
}

// LogAPI records an outbound API response to the database.
// If ScriptPath, ScriptFetchClass, ScriptFetchMethod, or ScriptLine are empty,
// they will be automatically resolved from the caller's call stack.
func (l *Logger) LogAPI(ctx context.Context, entry *ResponseAPI) error {
	if entry == nil {
		return errors.New("apilogger: entry cannot be nil")
	}

	// Auto-capture caller info if not supplied
	if entry.ScriptPath == "" && entry.ScriptFetchMethod == "" {
		caller := CaptureCaller(l.cfg.CallerSkip)
		entry.ScriptPath = caller.Path
		entry.ScriptFetchClass = caller.Class
		entry.ScriptFetchMethod = caller.Method
		entry.ScriptLine = caller.Line
	}

	// Auto-generate curl if missing and endpoint is provided
	if entry.Curl == "" && entry.Endpoint != "" && entry.HTTPMethod != "" {
		entry.Curl = BuildCurl(entry.HTTPMethod, entry.Endpoint, nil, entry.Param)
	}

	if l.cfg.Async {
		go func(e ResponseAPI) {
			if err := l.db.Table(l.cfg.APITableName).Create(&e).Error; err != nil {
				l.cfg.OnError(fmt.Errorf("apilogger: async LogAPI failed: %w", err))
			}
		}(*entry)
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return l.db.WithContext(ctx).Table(l.cfg.APITableName).Create(entry).Error
}

// LogCallback records an inbound callback/webhook event to the database.
func (l *Logger) LogCallback(ctx context.Context, entry *ResponseCallback) error {
	if entry == nil {
		return errors.New("apilogger: entry cannot be nil")
	}

	if l.cfg.Async {
		go func(e ResponseCallback) {
			if err := l.db.Table(l.cfg.CallbackTableName).Create(&e).Error; err != nil {
				l.cfg.OnError(fmt.Errorf("apilogger: async LogCallback failed: %w", err))
			}
		}(*entry)
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return l.db.WithContext(ctx).Table(l.cfg.CallbackTableName).Create(entry).Error
}

// DB returns the underlying gorm.DB instance.
func (l *Logger) DB() *gorm.DB {
	return l.db
}

// Config returns the active Logger configuration.
func (l *Logger) Config() Config {
	return l.cfg
}
