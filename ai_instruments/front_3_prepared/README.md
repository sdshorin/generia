# Generia Frontend Templates (Prepared)

Подготовленные шаблоны для нового фронтенда платформы Generia с улучшенной структурой, переиспользуемыми компонентами и чистым CSS.

## 📁 Структура проекта

```
front_3_prepared/
├── README.md                   # Этот файл
├── server.py                  # Python сервер для разработки
├── start-server.sh           # Bash скрипт для запуска сервера
│
├── CSS файлы:
├── common.css                # Общие стили и утилиты
├── components.css            # Стили для компонентов
├── main.css                  # Стили для главной страницы
├── catalog.css               # Стили для каталога миров
├── create-world.css          # Стили для страницы создания мира
├── feed.css                  # Стили для ленты постов
├── world-about.css           # Стили для страницы информации о мире
└── auth.css                  # Стили для страниц авторизации
│
├── components/               # Переиспользуемые компоненты
│   ├── header.html          # Общий хедер
│   ├── post-card.html       # Карточка поста
│   ├── world-card.html      # Карточка мира
│   ├── character-card.html  # Карточка персонажа
│   └── comment.html         # Компонент комментария
│
└── HTML страницы:
├── main.html                # Главная страница ✅
├── catalog.html             # Каталог миров ✅
├── create-world.html        # Создание мира ✅
├── world-feed.html          # Лента мира ✅
├── post-detail.html         # Детальный просмотр поста ✅
├── world-about.html         # Информация о мире ✅
├── login.html               # Страница входа ✅
└── register.html            # Страница регистрации ✅
```

## 🚀 Быстрый старт

### Способ 1: Bash скрипт (рекомендуется)

```bash
./start-server.sh
```

Или с указанием порта:

```bash
./start-server.sh 3000
```

### Способ 2: Python скрипт

```bash
python3 server.py
```

Или с указанием порта:

```bash
python3 server.py 3000
```

### Способ 3: Встроенный Python сервер

```bash
python3 -m http.server 8000
```

## 🌐 Доступные страницы

После запуска сервера открывайте в браузере:

- **Главная страница**: http://localhost:8000/main.html
- **Каталог миров**: http://localhost:8000/catalog.html
- **Создание мира**: http://localhost:8000/create-world.html
- **Лента мира**: http://localhost:8000/world-feed.html
- **Детальный просмотр поста**: http://localhost:8000/post-detail.html
- **Информация о мире**: http://localhost:8000/world-about.html
- **Вход в систему**: http://localhost:8000/login.html
- **Регистрация**: http://localhost:8000/register.html

## 📋 Что было сделано

### ✅ Структурные улучшения

1. **Выделены CSS стили:**
   - `common.css` - базовые стили, переменные, утилиты
   - `components.css` - стили для компонентов
   - Специфичные CSS файлы для каждой страницы

2. **Созданы переиспользуемые компоненты:**
   - Header с навигацией
   - Post Card для постов
   - World Card для миров
   - Character Card для персонажей
   - Comment component

3. **Переверстаны страницы:**
   - Главная страница с hero section и showcase
   - Каталог миров с фильтрами и поиском
   - Страница создания мира с анимированным прогрессом
   - Лента мира с минималистичным дизайном
   - Детальный просмотр поста с комментариями
   - Используются компоненты вместо дублирования кода

### ✅ Технические особенности

- **Отказ от Tailwind CSS** - все стили переписаны на обычный CSS
- **CSS Variables** - централизованные цвета, размеры, анимации
- **Модульная архитектура** - каждый компонент в отдельном файле
- **Адаптивный дизайн** - mobile-first подход
- **Простая навигация** - компоненты подгружаются через fetch API

## 🎨 CSS Архитектура

### Цветовая схема
```css
--color-primary: #2094f3        /* Основной синий */
--color-text-primary: #111518   /* Основной текст */
--color-text-secondary: #60778a /* Вторичный текст */
--color-bg-light: #f8f9fa       /* Светлый фон */
--color-border: #f0f2f5         /* Границы */
```

### Типографика
- **Шрифт**: Inter (Google Fonts)
- **Размеры**: От 12px (text-xs) до 48px (text-5xl)
- **Веса**: 400 (normal) до 900 (black)

### Компоненты
- `.btn` - базовая кнопка с модификаторами
- `.card` - базовая карточка с hover эффектами
- `.avatar` - аватары разных размеров
- `.input` - поля ввода с focus состояниями

## 🛠 Для разработчиков

### Добавление новой страницы

1. Создайте HTML файл в корне
2. Подключите CSS файлы:
   ```html
   <link rel="stylesheet" href="common.css">
   <link rel="stylesheet" href="components.css">
   <link rel="stylesheet" href="your-page.css">
   ```
3. Добавьте компоненты через fetch:
   ```javascript
   fetch('components/header.html')
     .then(res => res.text())
     .then(html => {
       document.getElementById('header-placeholder').innerHTML = html;
     });
   ```

### Создание нового компонента

1. Создайте HTML файл в папке `components/`
2. Добавьте стили в `components.css`
3. Используйте в страницах через fetch API

## ✅ Готовые страницы

- [x] `main.html` + `main.css` - Главная страница
- [x] `catalog.html` + `catalog.css` - Каталог миров
- [x] `create-world.html` + `create-world.css` - Создание мира
- [x] `world-feed.html` + `feed.css` - Лента мира
- [x] `post-detail.html` + дополнения к `feed.css` - Детальный просмотр поста
- [x] `world-about.html` + `world-about.css` - Информация о мире
- [x] `login.html` + `auth.css` - Страница входа
- [x] `register.html` + дополнения к `auth.css` - Страница регистрации

## 📝 TODO

Файлы, которые еще нужно создать:

- [ ] `character-profile.html` + стили - Профиль персонажа
- [ ] `settings.html` + стили - Настройки пользователя

## 🔧 Требования

- Python 3.x для запуска локального сервера
- Современный браузер с поддержкой ES6+
- Интернет для загрузки шрифтов Google Fonts

## 📞 Поддержка

Если у вас возникли проблемы:

1. Убедитесь, что Python 3 установлен: `python3 --version`
2. Проверьте, что порт свободен: `lsof -i :8000`
3. Запустите сервер из правильной директории

---

## 📖 Дополнительная документация

- **[DEVELOPMENT_GUIDE.md](DEVELOPMENT_GUIDE.md)** - Подробное техническое задание для продолжения разработки
- **[Исходное ТЗ](../front_3/generia_functional_spec.md)** - Функциональная спецификация всех страниц

---

**Статус**: Основные задачи выполнены ✅

**Выполнено на этом этапе**:
1. ✅ Создан `world-about.html` + `world-about.css` - Информация о мире
2. ✅ Созданы страницы авторизации: `login.html` и `register.html` + `auth.css`
3. ✅ Обновлен README с актуальной информацией

**Оставшиеся задачи** (низкий приоритет):
1. Создать профиль персонажа (`character-profile.html` + CSS)
2. Создать настройки пользователя (`settings.html` + CSS)

Подробный план работы см. в **DEVELOPMENT_GUIDE.md**