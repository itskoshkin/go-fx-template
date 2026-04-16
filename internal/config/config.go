package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/fx"

	"go-fx-template/pkg/minio"
	"go-fx-template/pkg/postgres"
	"go-fx-template/pkg/redis"
)

const Project = "go-fx-template"

const (
	FxStartUpTimeout  = 20 * time.Second
	FxShutDownTimeout = 20 * time.Second

	LogLevel       = "app.log.level"           // string ("DEBUG", "INFO", "WARN", "ERROR")
	LogFormat      = "app.log.log_format"      // string ("text" or "json")
	LogToConsole   = "app.log.log2console"     // bool
	LogToFile      = "app.log.log2file"        // bool
	LogFilePath    = "app.log.file_path"       // string (path)
	LogFileMode    = "app.log.file_mode"       // string ("append", "overwrite", "rotate")
	LogFilesFolder = "app.log.old_logs_folder" // string (path)

	GinReleaseMode            = "app.api.gin_release_mode"                // bool
	ApiHost                   = "app.api.host"                            // string
	ApiPort                   = "app.api.port"                            // int
	ApiBasePath               = "app.api.base_path"                       // string
	ApiShutdownTimeout        = "app.api.shutdown_timeout"                // time.Duration
	JwtIssuer                 = "app.api.auth.jwt_issuer"                 // string
	JwtAudience               = "app.api.auth.jwt_audience"               // string
	AccessTokenSecret         = "app.api.auth.access_token_secret"        // string
	RefreshTokenSecret        = "app.api.auth.refresh_token_secret"       // string
	AccessTokenTTL            = "app.api.auth.access_token_ttl"           // time.Duration
	RefreshTokenTTL           = "app.api.auth.refresh_token_ttl"          // time.Duration
	PasswordResetTokenTTL     = "app.api.auth.pwd_reset_token_ttl"        // time.Duration
	EmailVerificationTokenTTL = "app.api.auth.email_verify_token_ttl"     // time.Duration
	RequireEmailForUser       = "app.api.auth.require_email_for_user"     // bool
	RequireEmailVerification  = "app.api.auth.require_email_verification" // bool
	WebAppDomain              = "app.api.web_app.domain"                  // string

	GormStartupLogs           = "app.database.gorm.startup_logs"       // bool
	GormLogLevel              = "app.database.gorm.log_level"          // string ("SILENT", "ERROR", "WARN", "INFO", "DEBUG")
	DatabaseConnectionTimeout = "app.database.connection_timeout"      // time.Duration
	DatabaseHost              = "app.database.conn.host"               // string
	DatabasePort              = "app.database.conn.port"               // int
	DatabaseUser              = "app.database.conn.user"               // string
	DatabasePassword          = "app.database.conn.password"           // string
	DatabaseName              = "app.database.conn.database_name"      // string
	DatabaseSslMode           = "app.database.conn.ssl_mode"           // string
	DatabaseMaxOpenConn       = "app.database.pool.max_open_conns"     // int
	DatabaseMaxIdleConn       = "app.database.pool.max_idle_conns"     // int
	DatabaseConnMaxLifetime   = "app.database.pool.conn_max_lifetime"  // time.Duration
	DatabaseConnMaxIdleTime   = "app.database.pool.conn_max_idle_time" // time.Duration
	DatabaseAutoMigrate       = "app.database.auto_migrate"            // bool

	RedisHost     = "app.redis.host"     // string
	RedisPort     = "app.redis.port"     // int
	RedisPassword = "app.redis.password" // string
	RedisDB       = "app.redis.database" // int

	MinioEndpoint              = "app.minio.conn.endpoint"            // string
	MinioAccessKeyID           = "app.minio.conn.access_key_id"       // string
	MinioAccessKeySecret       = "app.minio.conn.access_key_secret"   // string
	MinioUseSSL                = "app.minio.conn.use_ssl"             // bool
	MinioBucketName            = "app.minio.conn.bucket_name"         // string
	MinioSetBucketPublicPolicy = "app.minio.set_bucket_public_policy" // bool
	MinioMaxFileSize           = "app.minio.max_upload_size_mb"       // int, MB

	EmailHost     = "app.email.host"     // string
	EmailPort     = "app.email.port"     // int
	EmailUser     = "app.email.user"     // string
	EmailPassword = "app.email.password" // string
	EmailFrom     = "app.email.from"     // string
)

func LoadConfig() {
	getEnv()
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		var osNotFound *fs.PathError
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); ok || errors.As(err, &osNotFound) {
			log.Printf("Warning: Config file not found, using defaults/env")
		} else {
			log.Fatalf("Error: Failed to read config: %v", err)
		}
	}
	applyDefaults()
	if err := validateConfigFields(); err != nil {
		log.Fatalf("Fatal: config validation error: %s", err)
	}
}

func getEnv() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	var binds = map[string]string{
		/* Logger   */ LogLevel: "LOG_LEVEL", LogFormat: "LOG_FORMAT", LogToConsole: "LOG_TO_CONSOLE", LogToFile: "LOG_TO_FILE", LogFilePath: "LOG_FILE_PATH", LogFileMode: "LOG_FILE_MODE", LogFilesFolder: "LOG_FILES_FOLDER",
		/* API      */ GinReleaseMode: "GIN_RELEASE_MODE", ApiHost: "APP_HOST", ApiPort: "APP_PORT", ApiBasePath: "API_BASE_PATH", ApiShutdownTimeout: "WEB_SERVER_SHUTDOWN_TIMEOUT", WebAppDomain: "WEB_APP_DOMAIN",
		/* JWT      */ JwtIssuer: "JWT_ISSUER", JwtAudience: "JWT_AUDIENCE", AccessTokenSecret: "JWT_ACCESS_TOKEN_SECRET", RefreshTokenSecret: "JWT_REFRESH_TOKEN_SECRET", AccessTokenTTL: "JWT_ACCESS_TOKEN_TTL", RefreshTokenTTL: "JWT_REFRESH_TOKEN_TTL", PasswordResetTokenTTL: "PWD_RESET_TOKEN_TTL", EmailVerificationTokenTTL: "EMAIL_VERIFY_TOKEN_TTL", RequireEmailForUser: "REQUIRE_EMAIL_FOR_USER", RequireEmailVerification: "REQUIRE_EMAIL_VERIFICATION",
		/* Database */ GormStartupLogs: "GORM_STARTUP_LOGS", GormLogLevel: "GORM_LOG_LEVEL", DatabaseConnectionTimeout: "DATABASE_CONNECTION_TIMEOUT", DatabaseHost: "DATABASE_HOST", DatabasePort: "DATABASE_PORT", DatabaseUser: "DATABASE_USER", DatabasePassword: "DATABASE_PASSWORD", DatabaseName: "DATABASE_NAME", DatabaseSslMode: "DATABASE_SSL_MODE", DatabaseMaxOpenConn: "DATABASE_MAX_OPEN_CONNS", DatabaseMaxIdleConn: "DATABASE_MAX_IDLE_CONNS", DatabaseConnMaxLifetime: "DATABASE_CONN_MAX_LIFETIME", DatabaseConnMaxIdleTime: "DATABASE_CONN_MAX_IDLE_TIME", DatabaseAutoMigrate: "DATABASE_AUTO_MIGRATE",
		/* Redis    */ RedisHost: "REDIS_HOST", RedisPort: "REDIS_PORT", RedisPassword: "REDIS_PASSWORD", RedisDB: "REDIS_DB",
		/* MinIO    */ MinioEndpoint: "MINIO_ENDPOINT", MinioAccessKeyID: "MINIO_ACCESS_KEY_ID", MinioAccessKeySecret: "MINIO_ACCESS_KEY_SECRET", MinioUseSSL: "MINIO_USE_SSL", MinioBucketName: "MINIO_BUCKET_NAME", MinioSetBucketPublicPolicy: "MINIO_SET_BUCKET_PUBLIC_POLICY", MinioMaxFileSize: "MINIO_MAX_FILE_SIZE",
		/* Email    */ EmailHost: "EMAIL_HOST", EmailPort: "EMAIL_PORT", EmailUser: "EMAIL_USER", EmailPassword: "EMAIL_PASSWORD", EmailFrom: "EMAIL_FROM",
	}
	for k, v := range binds {
		_ = viper.BindEnv(k, v)
	}
}

func applyDefaults() {
	var defaults = map[string]any{ // Will be set if not present, overwrites above required/dependent
		/* Logger   */ LogLevel: "INFO", LogFormat: "text", LogToConsole: true, LogToFile: true, LogFilePath: "application.log", LogFileMode: "append",
		/* API      */ GinReleaseMode: true, ApiHost: "0.0.0.0", ApiPort: 8080, ApiBasePath: "/api/v1", ApiShutdownTimeout: "5s",
		/* JWT      */ JwtIssuer: "go-fx-template", JwtAudience: "go-fx-template", AccessTokenTTL: "24h", RefreshTokenTTL: "168h", PasswordResetTokenTTL: "1h", EmailVerificationTokenTTL: "24h",
		/* Database */ DatabaseConnectionTimeout: "5s", DatabaseHost: "localhost", DatabasePort: 5432, DatabaseUser: "postgres", DatabaseName: "go-fx-template", DatabaseSslMode: "disable", GormLogLevel: "SILENT", GormStartupLogs: false,
		/* Redis    */ RedisHost: "localhost", RedisPort: 6379, RedisDB: 0,
		/* MinIO    */ MinioEndpoint: "localhost:9000", MinioUseSSL: false, MinioBucketName: "go-fx-template", MinioMaxFileSize: 10,
		/* Email    */ EmailHost: "smtp.gmail.com", EmailPort: 587,
	}
	for k, v := range defaults {
		if !viper.IsSet(k) || strings.TrimSpace(viper.GetString(k)) == "" {
			log.Printf("Warning: config field '%s' is missing or empty, defaulting to '%v'\n", k, v)
			viper.Set(k, v)
		}
	}
}

func validateConfigFields() error {
	var requiredFields = []string{ // Must be present and non-empty
		/* API      */ ApiHost, ApiPort, ApiBasePath, ApiShutdownTimeout,
		/* JWT      */ JwtIssuer, JwtAudience, AccessTokenSecret, RefreshTokenSecret, AccessTokenTTL, RefreshTokenTTL, PasswordResetTokenTTL, EmailVerificationTokenTTL,
		/* Database */ DatabaseHost, DatabasePort, DatabaseUser, DatabasePassword, DatabaseName, DatabaseSslMode,
		/* Redis    */ RedisHost, RedisPort,
		/* MinIO    */ MinioEndpoint, MinioAccessKeyID, MinioAccessKeySecret, MinioBucketName, MinioMaxFileSize,
		/* Email    */ EmailHost, EmailPort, EmailUser, EmailPassword, EmailFrom,
	}
	var dependentFields = map[string][]string{ // E.g. if A=true ==> must be non-empty B
		LogToFile: {LogFilePath},
	}
	var validatorRules = []validatorRule{
		{[]string{ApiPort, DatabasePort, RedisPort, EmailPort}, validatePort},                                                                                          // Ensure if provided port is in 1-65535 range
		{[]string{ApiShutdownTimeout, AccessTokenTTL, RefreshTokenTTL, PasswordResetTokenTTL, EmailVerificationTokenTTL, DatabaseConnectionTimeout}, validateDuration}, // Ensure if provided duration is > 0
		{[]string{LogLevel}, validateOneOf("DEBUG", "INFO", "WARN", "ERROR")},                                                                                          // If key is present, value must be one of these values
		{[]string{LogFormat}, validateOneOf("text", "json")},
		{[]string{LogFileMode}, validateOneOf("append", "overwrite", "rotate")},
		{[]string{DatabaseSslMode}, validateOneOf("disable", "require", "verify-ca", "verify-full")},
		{[]string{GormLogLevel}, validateOneOf("SILENT", "ERROR", "WARN", "INFO", "DEBUG")},
	}
	var conditionalRequired = map[string]map[string][]string{ // Validate that specific keys are present when a trigger key has a particular value
		LogFileMode: {
			"rotate": {LogFilesFolder},
		},
	}

	var missing []string
	for _, key := range requiredFields {
		if strings.TrimSpace(viper.GetString(key)) == "" {
			missing = append(missing, key)
		}
	}
	for triggerKey, requiredKeys := range dependentFields {
		if viper.GetBool(triggerKey) {
			for _, key := range requiredKeys {
				if strings.TrimSpace(viper.GetString(key)) == "" {
					missing = append(missing, fmt.Sprintf("%s (required when %s=true)", key, triggerKey))
				}
			}
		}
	}
	for key, cases := range conditionalRequired {
		val := strings.TrimSpace(viper.GetString(key))
		required, ok := cases[val]
		if !ok {
			continue
		}
		for _, field := range required {
			if strings.TrimSpace(viper.GetString(field)) == "" {
				missing = append(missing, fmt.Sprintf("%s (required when %s=%s)", field, key, val))
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields/values in config: %s", strings.Join(missing, ", "))
	}

	var invalid []string
	for _, rule := range validatorRules {
		for _, key := range rule.keys {
			val := strings.TrimSpace(viper.GetString(key))
			if val == "" {
				continue
			}
			if err := rule.validator(val); err != nil {
				invalid = append(invalid, fmt.Sprintf("%s: %s", key, err))
			}
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("invalid config values: %s", strings.Join(invalid, ", "))
	}

	return nil
}

type validatorRule struct {
	keys      []string
	validator func(string) error
}

func validateOneOf(allowed ...string) func(string) error {
	return func(v string) error {
		for _, a := range allowed {
			if v == a {
				return nil
			}
		}
		return fmt.Errorf("must be one of [%s], got %q", strings.Join(allowed, ", "), v)
	}
}

func validatePort(v string) error {
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 || n > 65535 {
		return fmt.Errorf("must be a port (1-65535), got %q", v)
	}
	return nil
}

func validateDuration(v string) error {
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fmt.Errorf("must be a positive duration, got %q", v)
	}
	return nil
}

func PostgresConfig() postgres.Config {
	return postgres.Config{
		Host:     viper.GetString(DatabaseHost),
		Port:     viper.GetString(DatabasePort),
		User:     viper.GetString(DatabaseUser),
		Password: viper.GetString(DatabasePassword),
		Database: viper.GetString(DatabaseName),
		SSLMode:  viper.GetString(DatabaseSslMode),

		MaxOpenConns:    viper.GetInt(DatabaseMaxOpenConn),
		MaxIdleConns:    viper.GetInt(DatabaseMaxIdleConn),
		ConnMaxLifetime: viper.GetDuration(DatabaseConnMaxLifetime),
		ConnMaxIdleTime: viper.GetDuration(DatabaseConnMaxIdleTime),

		ConnectionTimeout: viper.GetDuration(DatabaseConnectionTimeout),

		AutoMigrate: viper.GetBool(DatabaseAutoMigrate),
		StartupLogs: viper.GetBool(GormStartupLogs),
	}
}

func RedisConfig() redis.Config {
	return redis.Config{
		Addr:     viper.GetString(RedisHost),
		Port:     viper.GetString(RedisPort),
		Password: viper.GetString(RedisPassword),
		Database: viper.GetInt(RedisDB),
	}
}

func MinioConfig() minio.Config {
	return minio.Config{
		Endpoint:              viper.GetString(MinioEndpoint),
		AccessKeyID:           viper.GetString(MinioAccessKeyID),
		SecretAccessKey:       viper.GetString(MinioAccessKeySecret),
		UseSSL:                viper.GetBool(MinioUseSSL),
		BucketName:            viper.GetString(MinioBucketName),
		SetBucketPublicPolicy: viper.GetBool(MinioSetBucketPublicPolicy),
	}
}

var ProvideConfigs = fx.Module("configs",
	fx.Provide(PostgresConfig, RedisConfig, MinioConfig),
)
