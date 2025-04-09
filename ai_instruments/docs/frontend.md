# Фронтенд проекта Generia

## Обзор

Фронтенд Generia разработан с использованием React и TypeScript. Он обеспечивает удобный пользовательский интерфейс для взаимодействия с бэкенд-сервисами.

## Технологии

- **React 18+** - JavaScript-библиотека для создания пользовательских интерфейсов
- **TypeScript** - типизированный JavaScript для повышения надежности кода
- **React Router** - библиотека для маршрутизации в React-приложениях
- **Axios** - HTTP-клиент для выполнения запросов к API
- **Context API** - API React для управления глобальным состоянием приложения

## Структура фронтенда

```
/frontend
  /src
    /api
      - axios.ts         # Конфигурация HTTP-клиента
    /components
      - CreatePost.tsx   # Компонент создания поста
      - Feed.tsx         # Компонент отображения ленты
      - Login.tsx        # Компонент входа в систему
      - Navbar.tsx       # Компонент навигационной панели
      - Register.tsx     # Компонент регистрации
    /context
      - AuthContext.tsx  # Контекст для управления аутентификацией
    - App.tsx            # Корневой компонент приложения
    - index.tsx          # Точка входа в приложение
    - types.ts           # Типы TypeScript
  - package.json         # Зависимости и скрипты
  - tsconfig.json        # Конфигурация TypeScript
  - Dockerfile           # Инструкции для создания Docker-образа
  - nginx.conf           # Конфигурация Nginx
```

## Маршрутизация

В приложении реализованы следующие маршруты:

- `/` - Главная страница с лентой постов
- `/login` - Страница входа в систему
- `/register` - Страница регистрации
- `/create` - Страница создания поста

## Аутентификация

Аутентификация реализована с использованием JWT-токенов.

**AuthContext.tsx** управляет состоянием аутентификации и предоставляет следующие возможности:

- Вход пользователя
- Регистрация нового пользователя
- Выход из системы
- Проверка аутентификации пользователя
- Автоматическое добавление токена в заголовки запросов

## Компоненты

### Navbar.tsx

Навигационная панель, которая отображается на всех страницах приложения. Содержит ссылки на основные разделы и кнопку выхода из системы.

### Login.tsx

Форма входа в систему с валидацией полей и обработкой ошибок.

### Register.tsx

Форма регистрации нового пользователя с валидацией полей и обработкой ошибок.

### Feed.tsx

Компонент для отображения ленты постов. Обрабатывает пагинацию и подгрузку новых постов при прокрутке.

### CreatePost.tsx

Форма создания нового поста с возможностью загрузки изображений, добавления описания и отправки на сервер.

## Взаимодействие с API

Взаимодействие с бэкендом осуществляется через axios.ts, который настраивает HTTP-клиент Axios для работы с REST API.

```typescript
// axios.ts
import axios from 'axios';

// Use the API Gateway's address - using relative URL for better compatibility with proxy
const API_URL = '/api/v1';

const axiosInstance = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  // Add timeout to prevent hanging requests
  timeout: 10000,
});

// Intercept requests to add authorization token
axiosInstance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add response interceptor to handle common errors
axiosInstance.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // Log errors for debugging
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

export default axiosInstance;
```

## Основные типы данных

В файле `types.ts` определены основные типы данных, используемые в приложении:

```typescript
// types.ts
export interface User {
  id: string;
  username: string;
  email: string;
  profile_image?: string;
  bio?: string;
  created_at: string;
}

export interface Post {
  id: string;
  user_id: string;
  user?: User;
  caption: string;
  media_urls: string[];
  likes_count: number;
  comments_count: number;
  created_at: string;
}

export interface Comment {
  id: string;
  post_id: string;
  user: User;
  text: string;
  created_at: string;
}

export interface Like {
  user: User;
  created_at: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  loading: boolean;
  error: string | null;
}
```

## Сборка и запуск

Фронтенд запускается в Docker-контейнере с использованием Nginx в качестве веб-сервера.

```dockerfile
# Dockerfile
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

Nginx настроен для проксирования запросов к API на соответствующие бэкенд-сервисы:

```nginx
# nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/v1/ {
        proxy_pass http://api-gateway:8080/api/v1/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```
