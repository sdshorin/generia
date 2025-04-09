# Инфраструктура проекта Generia

## Обзор

Инфраструктура проекта Generia построена на базе Docker и Docker Compose, что обеспечивает легкость развертывания и масштабирования. Все компоненты системы упакованы в Docker-контейнеры и оркестрируются с помощью Docker Compose.

## Docker Compose

Все сервисы и зависимости управляются через Docker Compose. Основной файл конфигурации находится в корне проекта - `docker-compose.yml`.

```yaml
version: '3.8'

services:
  # База данных PostgreSQL
  postgres:
    image: postgres:14-alpine
    container_name: generia-postgres
    restart: always
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=generia
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/schema.sql:/docker-entrypoint-initdb.d/schema.sql
    networks:
      - generia_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  # Redis для кэширования
  redis:
    image: redis:alpine
    container_name: generia-redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - generia_network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
      
  # MongoDB для хранения данных взаимодействий
  mongodb:
    image: mongo:latest
    container_name: generia-mongodb
    restart: always
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=password
    volumes:
      - mongo_data:/data/db
    networks:
      - generia_network
    healthcheck:
      test: ["CMD", "mongosh", "--quiet", "--eval", "db.runCommand('ping').ok"]
      interval: 10s
      timeout: 10s
      retries: 5
      
  # MinIO для хранения медиафайлов
  minio:
    image: minio/minio
    container_name: generia-minio
    restart: always
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    networks:
      - generia_network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
      
  # Kafka для обмена событиями
  kafka:
    image: bitnami/kafka:latest
    container_name: generia-kafka
    restart: always
    ports:
      - "9092:9092"
    environment:
      - KAFKA_CFG_NODE_ID=1
      - KAFKA_CFG_PROCESS_ROLES=controller,broker
      - KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093
      - KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      - KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=1@kafka:9093
      - KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER
      - ALLOW_PLAINTEXT_LISTENER=yes
    volumes:
      - kafka_data:/bitnami/kafka
    networks:
      - generia_network

  # Service Discovery with Consul
  consul:
    image: consul:1.14
    ports:
      - "8500:8500"
    volumes:
      - consul_data:/consul/data
    command: "agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0"

  # Трассировка с Jaeger
  jaeger:
    image: jaegertracing/all-in-one:1.40
    ports:
      - "6831:6831/udp"
      - "16686:16686"

  # Prometheus для мониторинга
  prometheus:
    image: prom/prometheus:latest
    container_name: generia-prometheus
    restart: always
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - generia_network
      
  # Grafana для визуализации метрик
  grafana:
    image: grafana/grafana:latest
    container_name: generia-grafana
    restart: always
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - generia_network
    depends_on:
      - prometheus

  # API Gateway
  api-gateway:
    build:
      context: ./services/api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - PORT=8080

  # Auth Service
  auth-service:
    build:
      context: ./services/auth-service
    depends_on:
      - postgres
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - POSTGRES_URI=postgresql://generia:password@postgres:5432/generia?sslmode=disable
      - PORT=8081

  # Post Service
  post-service:
    build:
      context: ./services/post-service
    depends_on:
      - postgres
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - POSTGRES_URI=postgresql://generia:password@postgres:5432/generia?sslmode=disable
      - PORT=8082

  # Media Service
  media-service:
    build:
      context: ./services/media-service
    depends_on:
      - postgres
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - POSTGRES_URI=postgresql://generia:password@postgres:5432/generia?sslmode=disable
      - PORT=8083

  # Interaction Service
  interaction-service:
    build:
      context: ./services/interaction-service
    depends_on:
      - mongo
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - MONGO_URI=mongodb://mongo:27017/generia
      - PORT=8084

  # Feed Service
  feed-service:
    build:
      context: ./services/feed-service
    depends_on:
      - redis
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - REDIS_ADDR=redis:6379
      - PORT=8085

  # Cache Service
  cache-service:
    build:
      context: ./services/cache-service
    depends_on:
      - redis
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - REDIS_ADDR=redis:6379
      - PORT=8086

  # CDN Service
  cdn-service:
    build:
      context: ./services/cdn-service
    depends_on:
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - PORT=8087

  # Frontend application
  frontend:
    build:
      context: ./frontend
    ports:
      - "80:80"
    depends_on:
      - api-gateway

volumes:
  postgres_data:
  redis_data:
  mongo_data:
  minio_data:
  kafka_data:
  grafana_data:
```

## Docker

Каждый сервис упакован в Docker-контейнер с использованием соответствующего Dockerfile.

### API Gateway Dockerfile

```dockerfile
FROM golang:1.21-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
```

### Frontend Dockerfile

```dockerfile
FROM node:16-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Service Discovery with Consul

Consul используется для регистрации и обнаружения сервисов, что позволяет микросервисам находить друг друга без необходимости знать их точные адреса.

```go
// pkg/discovery/consul.go
package discovery

import (
    "github.com/hashicorp/consul/api"
)

type ServiceDiscovery interface {
    Register(name, host string, port int, tags []string) error
    Deregister() error
    GetService(name string) (string, error)
}

type ConsulClient struct {
    client *api.Client
    serviceID string
}

func NewConsulClient(address string) (*ConsulClient, error) {
    config := api.DefaultConfig()
    config.Address = address
    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }
    return &ConsulClient{client: client}, nil
}

// Реализация методов...
```

## Мониторинг с Prometheus

Prometheus используется для сбора и хранения метрик о работе сервисов. Конфигурация Prometheus находится в файле `configs/prometheus.yml`.

```yaml
# configs/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'api-gateway'
    scrape_interval: 5s
    static_configs:
      - targets: ['api-gateway:8080']

  - job_name: 'auth-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['auth-service:8081']

  - job_name: 'post-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['post-service:8082']

  - job_name: 'media-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['media-service:8083']

  - job_name: 'interaction-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['interaction-service:8084']

  - job_name: 'feed-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['feed-service:8085']

  - job_name: 'cache-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['cache-service:8086']

  - job_name: 'cdn-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['cdn-service:8087']
```

## Трассировка с Jaeger

Jaeger используется для распределенной трассировки запросов, что помогает понять поток запросов через различные микросервисы и обнаружить узкие места.

```go
// pkg/tracing/jaeger.go
package tracing

import (
    "io"

    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
)

// InitTracer создает новый трассировщик Jaeger
func InitTracer(serviceName, agentHostPort string) (opentracing.Tracer, io.Closer, error) {
    cfg := &config.Configuration{
        ServiceName: serviceName,
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:           true,
            LocalAgentHostPort: agentHostPort,
        },
    }
    tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
    if err != nil {
        return nil, nil, err
    }
    opentracing.SetGlobalTracer(tracer)
    return tracer, closer, nil
}
```

## Логирование

Для централизованного логирования используется пакет zap от Uber.

```go
// pkg/logger/logger.go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Logger обертка над zap.Logger
type Logger struct {
    *zap.Logger
}

// NewLogger создает новый логгер
func NewLogger(serviceName string, debug bool) (*Logger, error) {
    var config zap.Config
    if debug {
        config = zap.NewDevelopmentConfig()
    } else {
        config = zap.NewProductionConfig()
    }

    logger, err := config.Build()
    if err != nil {
        return nil, err
    }

    logger = logger.With(zap.String("service", serviceName))
    return &Logger{logger}, nil
}

// Sync синхронизирует буферы логгера
func (l *Logger) Sync() error {
    return l.Logger.Sync()
}
```

## Запуск и остановка

Для запуска всех сервисов используется команда:

```bash
docker-compose up -d
```

Для остановки всех сервисов:

```bash
docker-compose down
```

Для просмотра логов конкретного сервиса:

```bash
docker-compose logs -f <service-name>
```

## Масштабирование

Микросервисы могут быть масштабированы горизонтально с помощью Docker Compose:

```bash
docker-compose up -d --scale auth-service=3 --scale post-service=3
```

В реальном производственном окружении для оркестрации контейнеров лучше использовать Kubernetes или аналогичные инструменты.
