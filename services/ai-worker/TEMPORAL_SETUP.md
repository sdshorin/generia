# 🚀 Temporal AI Worker - Инструкции по запуску

Полная инструкция по запуску AI Worker с Temporal Workflow Engine.

## 📋 Предварительные требования

### 1. API ключи

Создайте файл `.env` в корне проекта:

```bash
# OpenRouter API для LLM
OPENROUTER_API_KEY=your_openrouter_api_key_here

# Runware API для генерации изображений  
RUNWARE_API_KEY=your_runware_api_key_here
```

### 2. Docker и Docker Compose

Убедитесь, что установлены:
- Docker >= 20.0
- Docker Compose >= 2.0

## 🐳 Запуск системы

### Шаг 1: Запуск базовой инфраструктуры

```bash
# Переходим в корень проекта
cd /Users/sergejsorin/study/diploma/generia

# Запускаем базовые сервисы
docker-compose up -d temporal temporal-web temporal-postgres mongodb minio consul
```

Проверьте что сервисы запустились:
```bash
docker-compose ps
```

### Шаг 2: Проверка Temporal

Откройте Temporal Web UI: http://localhost:8088

Если интерфейс доступен, Temporal готов к работе.

### Шаг 3: Запуск микросервисов (если нужно полное тестирование)

```bash
# Запуск основных микросервисов
docker-compose up -d postgres redis
docker-compose up -d auth-service character-service post-service media-service world-service api-gateway

# Проверяем статус
docker-compose ps
```

### Шаг 4: Запуск AI Worker

```bash
# Запуск AI Worker с Temporal
docker-compose up -d ai-worker

# Просмотр логов
docker-compose logs -f ai-worker
```

## 🧪 Тестирование

### Быстрый тест workflow

```bash
cd services/ai-worker

# Запуск тестового workflow
python test_temporal.py test
```

Тест создаст мир и запустит полную цепочку генерации. Результат можно отслеживать в:

- **Temporal Web UI**: http://localhost:8088
- **Логи AI Worker**: `docker-compose logs -f ai-worker`
- **MongoDB**: проверить данные в `generia_ai_worker.world_generation_status`

### Проверка статуса workflow

```bash
# Проверить статус по ID
python test_temporal.py status <workflow_id>

# Например:
python test_temporal.py status world-generation-test-world-20241228-143022
```

## 📊 Мониторинг

### Temporal Web UI (http://localhost:8088)

- **Workflows** → поиск по `world-generation-*`
- **Activities** → статистика выполнения
- **Task Queues** → загрузка workers

### Доступные Task Queues:

- `ai-worker-main` - основные workflows
- `ai-worker-llm` - LLM операции
- `ai-worker-images` - генерация изображений  
- `ai-worker-progress` - обновление прогресса
- `ai-worker-services` - gRPC вызовы

### Логи и метрики

```bash
# Логи AI Worker
docker-compose logs -f ai-worker

# Логи Temporal Server
docker-compose logs -f temporal

# Статус всех сервисов
docker-compose ps

# Использование ресурсов
docker stats
```

### MongoDB данные

```bash
# Подключение к MongoDB
docker exec -it generia-mongodb mongosh -u admin -p password

# Просмотр статуса генерации
use generia_ai_worker
db.world_generation_status.find().pretty()

# Просмотр параметров миров
db.world_parameters.find().pretty()
```

## 🔧 Настройка производительности

### Переменные окружения для AI Worker:

```bash
# Максимальная утилизация CPU (в docker-compose.yml)
MAX_CONCURRENT_TASKS=500
MAX_CONCURRENT_LLM_REQUESTS=50
MAX_CONCURRENT_IMAGE_REQUESTS=30
MAX_CONCURRENT_GRPC_CALLS=100
MAX_CONCURRENT_DB_OPERATIONS=20
MAX_WORKFLOW_TASKS_PER_WORKER=100
MAX_ACTIVITIES_PER_WORKER=200
```

### Workers конфигурация:

- **2 Main Workers** - workflows + general activities (load balancing)
- **1 LLM Worker** - специализируется на LLM запросах
- **1 Image Worker** - генерация изображений  
- **1 Progress Worker** - обновления в MongoDB
- **1 Service Worker** - gRPC вызовы

## 🐛 Диагностика проблем

### AI Worker не запускается

1. Проверьте зависимости:
```bash
docker-compose logs temporal
docker-compose logs mongodb
```

2. Проверьте API ключи в `.env`

3. Проверьте логи:
```bash
docker-compose logs ai-worker
```

### Workflows не выполняются

1. Проверьте Temporal Web UI: http://localhost:8088
2. Убедитесь что workers активны:
   - Temporal UI → Workers → Task Queues
3. Проверьте ошибки в Activities

### Долгое выполнение

1. Мониторьте прогресс в MongoDB:
```bash
db.world_generation_status.findOne({_id: "your_world_id"})
```

2. Проверьте загрузку API:
   - OpenRouter: ограничения rate limit
   - Runware: генерация изображений

### Ошибки в Activities

1. Temporal Web UI → Workflows → выберите workflow → History
2. Найдите failed activities и изучите ошибки
3. Проверьте retry policies

## 🔄 Развертывание изменений

### Обновление AI Worker:

```bash
# Пересборка и перезапуск
docker-compose build ai-worker
docker-compose up -d ai-worker

# Проверка логов
docker-compose logs -f ai-worker
```

### Масштабирование:

```bash
# Запуск дополнительных AI Workers
docker-compose up -d --scale ai-worker=3
```

## 📚 Полезные команды

### Temporal CLI (если установлен):

```bash
# Просмотр workflows
temporal workflow list

# Детали workflow
temporal workflow show -w <workflow_id>

# Отмена workflow
temporal workflow cancel -w <workflow_id>
```

### Docker управление:

```bash
# Остановка всех сервисов
docker-compose down

# Очистка данных (ОСТОРОЖНО!)
docker-compose down -v

# Перезапуск AI Worker
docker-compose restart ai-worker

# Просмотр ресурсов
docker system df
```

## 🎯 Ожидаемые результаты

При успешном тестировании вы увидите:

1. **Temporal Web UI** - активные workflows и completed activities
2. **MongoDB** - записи о прогрессе генерации
3. **MinIO** - сгенерированные изображения
4. **Логи AI Worker** - успешное выполнение задач

### Время выполнения (для тестового мира 3 пользователя, 10 постов):

- Инициализация: ~5 секунд
- Описание мира: ~30 секунд  
- Изображения мира: ~2 минуты
- Персонажи: ~3 минуты
- Посты: ~5 минут

**Общее время: ~10 минут**

## 🚨 Важные заметки

1. **API лимиты**: OpenRouter и Runware имеют rate limits
2. **Ресурсы**: генерация изображений требует времени
3. **Мониторинг**: всегда следите за Temporal Web UI
4. **Логи**: логи AI Worker содержат подробную информацию

---

✅ **Система готова к production с высокой производительностью и надежностью!**