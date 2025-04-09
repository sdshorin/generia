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
	Service  ServiceConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Consul   ConsulConfig
	Jaeger   JaegerConfig
	Redis    RedisConfig
	MongoDB  MongoDBConfig
	Minio    MinioConfig
	Kafka    KafkaConfig
	CDN      CDNConfig
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

// JaegerConfig holds Jaeger-related configuration
type JaegerConfig struct {
	Host string
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

	// Jaeger configuration
	jaegerHost := getEnv("JAEGER_HOST", "localhost:6831")

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
		Jaeger: JaegerConfig{
			Host: jaegerHost,
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