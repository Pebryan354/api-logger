package apilogger

// Config holds the configuration options for Logger.
type Config struct {
	APITableName      string
	CallbackTableName string
	AutoMigrate       bool
	Async             bool
	CallerSkip        int
	OnError           func(err error)
}

// Option configures the Logger.
type Option func(*Config)

// WithAPITableName sets a custom table name for API response logs.
func WithAPITableName(tableName string) Option {
	return func(c *Config) {
		if tableName != "" {
			c.APITableName = tableName
		}
	}
}

// WithCallbackTableName sets a custom table name for callback logs.
func WithCallbackTableName(tableName string) Option {
	return func(c *Config) {
		if tableName != "" {
			c.CallbackTableName = tableName
		}
	}
}

// WithAutoMigrate toggles automatic table creation on initialization.
func WithAutoMigrate(enabled bool) Option {
	return func(c *Config) {
		c.AutoMigrate = enabled
	}
}

// WithAsync enables non-blocking asynchronous log insertion.
func WithAsync(enabled bool) Option {
	return func(c *Config) {
		c.Async = enabled
	}
}

// WithCallerSkip sets the stack frame depth to skip when capturing caller info.
func WithCallerSkip(skip int) Option {
	return func(c *Config) {
		c.CallerSkip = skip
	}
}

// WithErrorHandler sets a callback function for errors (especially useful for async mode).
func WithErrorHandler(fn func(err error)) Option {
	return func(c *Config) {
		c.OnError = fn
	}
}
