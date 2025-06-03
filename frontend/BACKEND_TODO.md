# Backend TODO: Отсутствующие данные для нового фронтенда

## 🚨 КРИТИЧЕСКИЕ ДАННЫЕ (нужны для основной функциональности)

### World Model - Дополнительные поля

**Текущие поля в API**: ✅
```typescript
interface World {
  id: string;
  name: string;
  description?: string;
  prompt: string;
  creator_id?: string;
  generation_status: string;
  status: string;
  users_count: number;
  posts_count: number;
  created_at: string;
  updated_at: string;
  is_joined?: boolean;
  is_active?: boolean;
  image_url?: string;
  icon_url?: string;
}
```

**НУЖНО ДОБАВИТЬ**: ❌
```typescript
interface World {
  // ... существующие поля
  
  // Для карточек мира
  cover_image_url: string;          // Обложка мира (отличается от image_url)
  icon_url: string;                 // Иконка мира (64x64)
  likes_count: number;              // Общее количество лайков в мире
  
  // Для детальной страницы мира (/worlds/:id/about)
  detailed_description: string;     // Расширенное описание
  world_characteristics: {
    technology_level: number;       // 0-100
    magic_presence: number;         // 0-100  
    social_structure: number;       // 0-100
    geography_diversity: number;    // 0-100
  };
  history_timeline: Array<{
    period: string;                 // "Ancient Era", "Modern Times"
    events: string[];              // Список событий
  }>;
  featured_character_ids: string[]; // ID топ персонажей мира
}
```

### Character Model - Новые поля

**НУЖНО ДОБАВИТЬ**: ❌
```typescript
interface Character {
  id: string;
  name: string;
  world_id: string;
  avatar_url: string;
  
  // Новые поля для профиля персонажа
  role: string;                     // "Merchant", "Warrior", "Mage"
  biography: string;                // Детальная биография
  traits: string[];                // ["Brave", "Intelligent", "Mysterious"]
  specializations: string[];       // ["Combat", "Magic", "Trading"]
  posts_count: number;             // Количество постов персонажа
  likes_received: number;          // Общее количество лайков
  created_at: string;
  is_ai: boolean;
}
```

### User Model - Дополнительные поля

**НУЖНО ДОБАВИТЬ**: ❌
```typescript
interface User {
  // ... существующие поля
  
  // Для хедера и настроек
  credits_balance: number;          // Баланс кредитов пользователя
  avatar_url?: string;             // Аватар пользователя
  
  // Для страницы настроек
  notification_settings: {
    email_notifications: boolean;
    push_notifications: boolean;
    world_updates: boolean;
    character_interactions: boolean;
  };
  privacy_settings: {
    profile_visibility: "public" | "private";
    activity_tracking: boolean;
    data_sharing: boolean;
  };
}
```

---

## 📊 НОВЫЕ API ENDPOINTS

### 1. Credits & Billing

**НУЖНО ДОБАВИТЬ**: ❌
```
GET /api/v1/user/credits                    # Получить баланс кредитов
POST /api/v1/user/credits/purchase          # Купить кредиты
GET /api/v1/user/transactions               # История транзакций
```

### 2. World Details

**НУЖНО ДОБАВИТЬ**: ❌
```
GET /api/v1/worlds/{id}/about               # Детальная информация о мире
GET /api/v1/worlds/{id}/characters          # Топ персонажи мира
GET /api/v1/worlds/{id}/statistics          # Статистика мира
```

### 3. Character Profiles  

**НУЖНО ДОБАВИТЬ**: ❌
```
GET /api/v1/characters/{id}/profile         # Полный профиль персонажа
GET /api/v1/characters/{id}/posts           # Посты персонажа
GET /api/v1/worlds/{worldId}/characters/{id} # Персонаж в контексте мира
```

### 4. User Settings

**НУЖНО ДОБАВИТЬ**: ❌
```
GET /api/v1/user/settings                   # Получить настройки
PUT /api/v1/user/settings                   # Обновить настройки
PUT /api/v1/user/avatar                     # Обновить аватар
DELETE /api/v1/user/account                 # Удалить аккаунт
```

---

## 🎨 МЕДИА И АССЕТЫ

### World Assets
**НУЖНО ДОБАВИТЬ**: ❌
- Генерация `cover_image_url` (широкоформатные обложки 16:9)
- Генерация `icon_url` (квадратные иконки 64x64)
- Разделение между основным изображением и обложкой

### Character Assets  
**НУЖНО ДОБАВИТЬ**: ❌
- Система генерации аватаров для AI персонажей
- Различные размеры аватаров (32px, 64px, 128px)

---

## 🔄 МОДИФИКАЦИИ СУЩЕСТВУЮЩИХ ENDPOINTS

### World Creation
**НУЖНО МОДИФИЦИРОВАТЬ**: `/api/v1/worlds` (POST)

**Текущий запрос**:
```json
{
  "name": "string",
  "description": "string", 
  "prompt": "string",
  "charactersCount": "number",
  "postsCount": "number"
}
```

**НУЖНО ДОБАВИТЬ**: ❌
```json
{
  // ... существующие поля
  "generate_cover_image": true,        // Генерировать обложку
  "generate_icon": true               // Генерировать иконку
}
```

### Post Creation
**НУЖНО МОДИФИЦИРОВАТЬ**: `/api/v1/worlds/{worldId}/posts` (POST)

**ДОБАВИТЬ в ответ**: ❌
```json
{
  // ... существующие поля поста
  "character": {
    "id": "string",
    "name": "string", 
    "avatar_url": "string",
    "role": "string"
  }
}
```

---

## 📱 ИНТЕГРАЦИИ И СЕРВИСЫ

### Credits System
**НУЖНО СОЗДАТЬ**: ❌
- Система кредитов с транзакциями
- Интеграция с платежной системой
- Расчет стоимости генерации контента

### Notification System  
**НУЖНО СОЗДАТЬ**: ❌
- Email уведомления
- Push уведомления (если потребуется)
- Настройки уведомлений в профиле

---

## 🚀 ПРИОРИТЕТЫ РЕАЛИЗАЦИИ

### Высокий приоритет (для основного функционала)
1. ✅ World: `cover_image_url`, `icon_url`, `likes_count`
2. ✅ Character: базовые поля для профилей
3. ✅ User: `credits_balance`, `avatar_url`
4. ✅ Credits API endpoints

### Средний приоритет (для расширенного функционала)
1. ⭕ World: детальная информация (`world_characteristics`, `history_timeline`)
2. ⭕ Character: расширенные поля (`biography`, `traits`, `specializations`)
3. ⭕ User settings API

### Низкий приоритет (можно отложить)
1. 🔵 Notification system
2. 🔵 Privacy settings
3. 🔵 Advanced analytics

---

## 📝 ВРЕМЕННЫЕ ЗАГЛУШКИ

**Пока бэкенд не готов, фронтенд будет использовать заглушки из `mockData.ts`**:

- Характеристики миров (technology_level, magic_presence, etc.)
- История миров (timeline)
- Биографии персонажей  
- Баланс кредитов (фиксированное значение 1,250)
- Настройки пользователя

---

## ✅ СТАТУС ОТСЛЕЖИВАНИЯ

**Обновлять этот раздел при добавлении полей в бэкенд**

- [ ] World: cover_image_url, icon_url, likes_count
- [ ] Character: role, biography, traits, posts_count
- [ ] User: credits_balance, avatar_url
- [ ] Credits API endpoints
- [ ] World details API
- [ ] Character profile API
- [ ] User settings API

**Последнее обновление**: 06/03/2025