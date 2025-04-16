package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	Service      ServiceConfig
	Database     DatabaseConfig
	JWT          JWTConfig
	Consul       ConsulConfig
	Telemetry    TelemetryConfig
	Redis        RedisConfig
	MongoDB      MongoDBConfig
	Minio        MinioConfig
	Kafka        KafkaConfig
	CDN          CDNConfig
	Jaeger       JaegerConfig
}

// ServiceConfig holds service-related configuration
type ServiceConfig struct {
	Name string
	Host string
	Port int
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// ConsulConfig holds Consul-related configuration
type ConsulConfig struct {
	Address string
}

// TelemetryConfig holds OpenTelemetry configuration
type TelemetryConfig struct {
	Endpoint        string // OTLP endpoint
	ServiceName     string
	Environment     string
	SamplingRatio   float64
	PropagatorType  string // "b3", "w3c", "all"
	DisableMetrics  bool
	DisableTracing  bool
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// MongoDBConfig holds MongoDB-related configuration
type MongoDBConfig struct {
	URI      string
	Database string
}

// MinioConfig holds Minio-related configuration
type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// KafkaConfig holds Kafka-related configuration
type KafkaConfig struct {
	Brokers []string
}

// CDNConfig holds CDN-related configuration
type CDNConfig struct {
	Domain     string
	DefaultTTL int
	SigningKey string
}

// JaegerConfig holds Jaeger-related configuration
type JaegerConfig struct {
	Host string
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Service configuration
	serviceName := getEnv("SERVICE_NAME", "app")
	serviceHost := getEnv("SERVICE_HOST", "0.0.0.0")
	servicePortStr := getEnv("SERVICE_PORT", "8080")
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid service port: %s", servicePortStr)
	}

	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "generia")
	dbSSLMode := getEnv("DB_SSL_MODE", "disable")

	// JWT configuration
	jwtSecret := getEnv("JWT_SECRET", "your_jwt_secret_key")
	jwtExpirationStr := getEnv("JWT_EXPIRATION", "24h")
	jwtExpiration, err := time.ParseDuration(jwtExpirationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT expiration: %s", jwtExpirationStr)
	}

	// Consul configuration
	consulAddress := getEnv("CONSUL_ADDRESS", "localhost:8500")

	// OpenTelemetry configuration
	otlpEndpoint := getEnv("OTLP_ENDPOINT", "jaeger:4318") // Default to Jaeger OTLP endpoint
	telemetryServiceName := getEnv("TELEMETRY_SERVICE_NAME", serviceName)
	telemetryEnvironment := getEnv("TELEMETRY_ENVIRONMENT", "production")
	samplingRatioStr := getEnv("TELEMETRY_SAMPLING_RATIO", "1.0")
	samplingRatio, err := strconv.ParseFloat(samplingRatioStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid sampling ratio: %s", samplingRatioStr)
	}
	propagatorType := getEnv("TELEMETRY_PROPAGATOR", "w3c")
	disableMetricsStr := getEnv("TELEMETRY_DISABLE_METRICS", "false")
	disableMetrics, err := strconv.ParseBool(disableMetricsStr)
	if err != nil {
		return nil, fmt.Errorf("invalid disable metrics flag: %s", disableMetricsStr)
	}
	disableTracingStr := getEnv("TELEMETRY_DISABLE_TRACING", "false")
	disableTracing, err := strconv.ParseBool(disableTracingStr)
	if err != nil {
		return nil, fmt.Errorf("invalid disable tracing flag: %s", disableTracingStr)
	}

	// Redis configuration
	redisAddress := getEnv("REDIS_ADDRESS", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDBStr := getEnv("REDIS_DB", "0")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis DB: %s", redisDBStr)
	}

	// MongoDB configuration
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	mongoDatabase := getEnv("MONGODB_DATABASE", "generia")

	// Minio configuration
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	minioBucket := getEnv("MINIO_BUCKET", "generia-images")
	minioUseSSLStr := getEnv("MINIO_USE_SSL", "false")
	minioUseSSL, err := strconv.ParseBool(minioUseSSLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Minio SSL flag: %s", minioUseSSLStr)
	}

	// Kafka configuration
	kafkaBrokersStr := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaBrokers := []string{kafkaBrokersStr}

	// CDN configuration
	cdnDomain := getEnv("CDN_DOMAIN", "localhost")
	cdnDefaultTTLStr := getEnv("CDN_DEFAULT_TTL", "86400")
	cdnDefaultTTL, err := strconv.Atoi(cdnDefaultTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CDN default TTL: %s", cdnDefaultTTLStr)
	}
	cdnSigningKey := getEnv("CDN_SIGNING_KEY", "your_cdn_signing_key")
	
	// Jaeger configuration
	jaegerHost := getEnv("JAEGER_HOST", "jaeger")
	jaegerPort := getEnv("JAEGER_PORT", "6831")

	return &Config{
		Service: ServiceConfig{
			Name: serviceName,
			Host: serviceHost,
			Port: servicePort,
		},
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
			SSLMode:  dbSSLMode,
		},
		JWT: JWTConfig{
			Secret:     jwtSecret,
			Expiration: jwtExpiration,
		},
		Consul: ConsulConfig{
			Address: consulAddress,
		},
		Telemetry: TelemetryConfig{
			Endpoint:       otlpEndpoint,
			ServiceName:    telemetryServiceName,
			Environment:    telemetryEnvironment,
			SamplingRatio:  samplingRatio,
			PropagatorType: propagatorType,
			DisableMetrics: disableMetrics,
			DisableTracing: disableTracing,
		},
		Redis: RedisConfig{
			Address:  redisAddress,
			Password: redisPassword,
			DB:       redisDB,
		},
		MongoDB: MongoDBConfig{
			URI:      mongoURI,
			Database: mongoDatabase,
		},
		Minio: MinioConfig{
			Endpoint:  minioEndpoint,
			AccessKey: minioAccessKey,
			SecretKey: minioSecretKey,
			Bucket:    minioBucket,
			UseSSL:    minioUseSSL,
		},
		Kafka: KafkaConfig{
			Brokers: kafkaBrokers,
		},
		CDN: CDNConfig{
			Domain:     cdnDomain,
			DefaultTTL: cdnDefaultTTL,
			SigningKey: cdnSigningKey,
		},
		Jaeger: JaegerConfig{
			Host: jaegerHost,
			Port: jaegerPort,
		},
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}