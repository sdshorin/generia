# Generia Frontend Development Guide

**ГОТОВОЕ ТЕХНИЧЕСКОЕ ЗАДАНИЕ** для переноса подготовленного фронтенда на реальный сайт Generia.

Все 10 страниц созданы и готовы к переносу. Данный документ содержит полную спецификацию компонентов, стилей и зависимостей между файлами.

---

## 📁 Архитектура готового проекта

### Структура файлов

```
front_3_prepared/
├── README.md                   # Краткое описание проекта
├── DEVELOPMENT_GUIDE.md        # Этот файл - полное ТЗ
├── server.py                  # Python сервер для тестирования
├── start-server.sh           # Bash скрипт для запуска сервера
│
├── 🎨 CSS ФАЙЛЫ:
├── common.css                # Базовые стили и утилиты (700+ строк) ✅
├── components.css            # Стили для переиспользуемых компонентов (400+ строк) ✅ 
├── main.css                  # Стили для главной страницы (300+ строк) ✅
├── catalog.css               # Стили для каталога миров (250+ строк) ✅
├── create-world.css          # Стили для страницы создания мира (700+ строк) ✅
├── feed.css                  # Стили для ленты постов и детального просмотра (480+ строк) ✅
├── world-about.css           # Стили для страницы информации о мире (500+ строк) ✅
├── auth.css                  # Стили для страниц авторизации (400+ строк) ✅
├── character-profile.css     # Стили для профиля персонажа (400+ строк) ✅
├── settings.css              # Стили для настроек пользователя (400+ строк) ✅
│
├── 🧩 КОМПОНЕНТЫ (components/):
├── header.html              # Общий хедер с навигацией ✅
├── post-card.html           # Карточка поста для ленты ✅
├── world-card.html          # Карточка мира для каталога ✅
├── character-card.html      # Карточка персонажа ✅
├── comment.html             # Компонент комментария ✅
│
└── 📄 HTML СТРАНИЦЫ:
├── main.html                # ✅ Главная страница
├── catalog.html             # ✅ Каталог миров
├── create-world.html        # ✅ Создание мира
├── world-feed.html          # ✅ Лента мира
├── post-detail.html         # ✅ Детальный просмотр поста
├── world-about.html         # ✅ Информация о мире
├── login.html               # ✅ Страница входа
├── register.html            # ✅ Страница регистрации
├── character-profile.html   # ✅ Профиль персонажа
└── settings.html            # ✅ Настройки пользователя
```

---

## 🎨 Описание CSS файлов

### `common.css` (700+ строк) - БАЗОВАЯ СИСТЕМА СТИЛЕЙ
**Назначение**: Основа всех стилей проекта  
**Содержимое**:
- CSS переменные (цвета, размеры, шрифты, отступы, тени, анимации)
- CSS Reset и базовая типографика
- Утилитарные классы (layout, spacing, colors, typography)
- Базовые компоненты (.btn, .card, .avatar, .input, .form-*)
- Универсальные контейнеры (.container, .min-h-screen, .flex-col)
- Анимации и переходы (pulse, spin, transitions)
- Адаптивные медиа-запросы (mobile-first подход)

### `components.css` (400+ строк) - КОМПОНЕНТЫ
**Назначение**: Стили для переиспользуемых компонентов  
**Содержимое**:
- `.header` - стили хедера и навигации с мобильным меню
- `.post-card` - стили карточки поста с hover эффектами
- `.world-card` - стили карточки мира с статистикой
- `.character-card` - стили карточки персонажа
- `.comment` - стили компонента комментария с действиями
- Мобильные меню и адаптация для всех компонентов

### Специфичные CSS файлы страниц:

#### `main.css` (300+ строк) - ГЛАВНАЯ СТРАНИЦА
- Hero section с градиентным фоном
- Explore section с интерактивными мирами
- World showcase с переключением миров
- "How it works" секция с иконками
- Примеры постов и CTA кнопки

#### `catalog.css` (250+ строк) - КАТАЛОГ МИРОВ
- Фильтры и поиск в хедере каталога
- Сетка миров с адаптивными колонками
- Активные состояния фильтров
- Hover эффекты для карточек миров

#### `create-world.css` (700+ строк) - СОЗДАНИЕ МИРА
- Слайдеры (range input) с кастомными стилями
- Форма создания мира с текстовыми полями
- Отображение стоимости в кредитах
- Прогресс генерации с 6 этапами
- Анимированные прогресс-бары для персонажей и постов
- Два состояния: форма создания + экран прогресса

#### `feed.css` (480+ строк) - ЛЕНТА И ДЕТАЛЬНЫЙ ПРОСМОТР
- Минималистичный хедер для ленты
- World header с обложкой мира и overlay
- Feed контейнер с ограниченной шириной
- Стили детального просмотра поста
- Секция комментариев с формой добавления
- Load more кнопка и infinite scroll

#### `world-about.css` (500+ строк) - ИНФОРМАЦИЯ О МИРЕ
- World cover section с overlay и информацией
- Двухколоночная сетка (детали мира + статистика)
- History timeline с этапами развития мира
- World characteristics с прогресс-барами
- Statistics section со списком показателей
- Characters list с прокруткой и hover эффектами

#### `auth.css` (400+ строк) - АВТОРИЗАЦИЯ
- Центрированный layout без хедера
- Auth card с белым фоном и тенью
- Стили форм с полями ввода
- Валидация полей (error, success states)
- Password requirements индикатор
- Checkbox стили для "Remember me" и "Terms"

#### `character-profile.css` (400+ строк) - ПРОФИЛЬ ПЕРСОНАЖА
- Character header с большим аватаром и информацией
- Statistics section с постами, лайками, мирами
- Posts gallery с использованием post-card стилей
- About section с описанием персонажа
- World affiliation секция
- Traits grid с характеристиками и специализациями

#### `settings.css` (400+ строк) - НАСТРОЙКИ ПОЛЬЗОВАТЕЛЯ
- Settings navigation (табы или меню)
- Form sections стили
- Settings cards с группировкой опций
- Credits display и управление балансом
- Switch/toggle компоненты для настроек
- Transaction history и privacy controls

---

## 🧩 Описание компонентов

### `components/header.html` - ОБЩИЙ ХЕДЕР
**Функции**:
- Логотип с переходом на главную
- Навигация (Home, Explore Worlds, Feed, Research Paper, Create World)
- Мобильное меню с тем же набором ссылок
- Отображение кредитов пользователя (адаптивно)
- Аватар пользователя
- Автоматическая установка активной ссылки

**JavaScript функции**:
- `toggleMobileMenu()` - переключение мобильного меню
- `setActiveNavLink()` - установка активной ссылки на основе текущей страницы

**Стили**: Использует классы из `components.css` (.header, .nav, .mobile-menu)

### `components/post-card.html` - КАРТОЧКА ПОСТА
**Функции**:
- Аватар автора с переходом к профилю персонажа
- Информация об авторе (имя, роль, время публикации)
- Изображение поста с переходом к детальному просмотру
- Действия (лайк, комментарий, сохранение)
- Счетчики лайков и комментариев
- Подпись к посту

**JavaScript функции**:
- `toggleLike(button)` - переключение лайка с анимацией
- `toggleSave(button)` - сохранение поста
- `goToPost(postId)` - переход к детальному просмотру
- `goToCharacter(characterId)` - переход к профилю персонажа

**Стили**: Использует классы из `components.css` (.post-card, .post-header, .post-actions)

### `components/world-card.html` - КАРТОЧКА МИРА
**Функции**:
- Обложка мира с overlay для иконки
- Иконка мира в углу
- Название и описание мира
- Статистика (персонажи, посты, лайки)
- Кнопка входа в мир
- Поддержка бейджей (Popular, New)

**JavaScript функции**:
- `goToWorld(worldId)` - переход в ленту мира

**Стили**: Использует классы из `components.css` (.world-card, .world-card-image)

### `components/character-card.html` - КАРТОЧКА ПЕРСОНАЖА
**Функции**:
- Аватар персонажа
- Имя и роль персонажа
- Статистика постов и лайков
- Переход к профилю персонажа

**JavaScript функции**:
- `goToCharacter(characterId)` - переход к профилю персонажа

**Стили**: Использует классы из `components.css` (.character-card)

### `components/comment.html` - КОМПОНЕНТ КОММЕНТАРИЯ
**Функции**:
- Аватар автора комментария
- Информация об авторе и времени
- Текст комментария
- Действия (лайк, ответ)
- Поле для ответа (показывается при нажатии)

**JavaScript функции**:
- `toggleCommentLike(button)` - лайк комментария
- `showReplyInput(commentId)` - показать поле ответа
- `postReply(commentId)` - отправить ответ

**Стили**: Использует классы из `components.css` (.comment, .comment-actions)

---

## 📄 Детальное описание страниц

### 1. ✅ Главная страница (`main.html`)

**Исходный макет**: `ai_instruments/front_3/main.html`  
**Готовый файл**: `ai_instruments/front_3_prepared/main.html`  
**Стили**: `common.css` + `components.css` + `main.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/main.html
- ai_instruments/front_3_prepared/main.css
- ai_instruments/front_3_prepared/common.css
- ai_instruments/front_3_prepared/components.css
- ai_instruments/front_3_prepared/components/header.html
- ai_instruments/front_3_prepared/components/post-card.html

ДЛЯ РЕФЕРЕНСА (исходный макет):
- ai_instruments/front_3/main.html
```

**Функциональность**:
- Hero section с призывом к действию и переходом к созданию мира
- Интерактивный блок топ-5 миров с переключением
- Превью выбранного мира с описанием и статистикой
- Секция "How it works" с техническими деталями
- Примеры постов из мира
- CTA для создания мира

**Компоненты**: header, post-card (x2)

---

### 2. ✅ Каталог миров (`catalog.html`)

**Исходный макет**: `ai_instruments/front_3/worlds_catalog_page.html`  
**Готовый файл**: `ai_instruments/front_3_prepared/catalog.html`  
**Стили**: `common.css` + `components.css` + `catalog.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/catalog.html
- ai_instruments/front_3_prepared/catalog.css
- ai_instruments/front_3_prepared/common.css
- ai_instruments/front_3_prepared/components.css
- ai_instruments/front_3_prepared/components/header.html
- ai_instruments/front_3_prepared/components/world-card.html

ДЛЯ РЕФЕРЕНСА (исходный макет):
- ai_instruments/front_3/worlds_catalog_page.html
```

**Функциональность**:
- Сетка карточек миров (6 карточек)
- Загрузка дополнительных миров

**Компоненты**: header, world-card (x6)

---

### 3. ✅ Создание мира (`create-world.html`)

**Исходный макет**: `ai_instruments/front_3/create_world_page.html`  
**Готовый файл**: `ai_instruments/front_3_prepared/create-world.html`  
**Стили**: `common.css` + `components.css` + `create-world.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/create-world.html
- ai_instruments/front_3_prepared/create-world.css
- ai_instruments/front_3_prepared/common.css
- ai_instruments/front_3_prepared/components.css
- ai_instruments/front_3_prepared/components/header.html

ДЛЯ РЕФЕРЕНСА (исходный макет):
- ai_instruments/front_3/create_world_page.html
```

**Функциональность**:
- Форма создания мира с текстовым полем (textarea)
- Слайдеры для настройки количества персонажей (5-50)
- Слайдеры для настройки количества постов (10-100)
- Отображение стоимости в кредитах с разбивкой
- Прогресс генерации с 6 этапами (параллельное выполнение)
- Прогресс-бары для персонажей и постов
- Переход в созданный мир после завершения

**Компоненты**: header  
**Особенности**: Два состояния (форма + прогресс), JavaScript анимации

---

### 4. ✅ Лента мира (`world-feed.html`)

**Исходный макет**: `ai_instruments/front_3/world_feed_page.html`  
**Готовый файл**: `ai_instruments/front_3_prepared/world-feed.html`  
**Стили**: `common.css` + `components.css` + `feed.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/world-feed.html
- ai_instruments/front_3_prepared/feed.css
- ai_instruments/front_3_prepared/common.css
- ai_instruments/front_3_prepared/components.css
- ai_instruments/front_3_prepared/components/post-card.html

ДЛЯ РЕФЕРЕНСА (исходный макет):
- ai_instruments/front_3/world_feed_page.html
```

**Функциональность**:
- Минимальный хедер с back кнопкой и кредитами
- World header с обложкой мира и информацией
- Кнопка перехода к деталям мира
- Лента постов с разными авторами (4 поста)
- Load more кнопка и infinite scroll
- Взаимодействия с постами

**Компоненты**: post-card (x4, с разными данными)  
**Особенность**: Минималистичный дизайн без обычного хедера, собственный feed-header

---

### 5. ✅ Детальный просмотр поста (`post-detail.html`)

**Исходный макет**: `ai_instruments/front_3/post_detail_page.html`  
**Готовый файл**: `ai_instruments/front_3_prepared/post-detail.html`  
**Стили**: `common.css` + `components.css` + `feed.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/post-detail.html
- ai_instruments/front_3_prepared/feed.css
- ai_instruments/front_3_prepared/common.css
- ai_instruments/front_3_prepared/components.css
- ai_instruments/front_3_prepared/components/comment.html

ДЛЯ РЕФЕРЕНСА (исходный макет):
- ai_instruments/front_3/post_detail_page.html
```

**Функциональность**:
- Минимальный хедер с back кнопкой (как в feed)
- Полноразмерное отображение поста
- Расширенная информация об авторе
- Секция комментариев с загрузкой через fetch
- Поле добавления комментария с auto-resize
- Действия с постом (лайк, сохранение)

**Компоненты**: comment (x3, с разными данными)  
**Особенности**: Использует стили из feed.css, динамическая загрузка комментариев

---

### 6. ✅ Информация о мире (`world-about.html`)

**Исходный макет**: `ai_instruments/front_3/world_about_page.html`  
**Готовый файл**: `ai_instruments/front_3_prepared/world-about.html`  
**Стили**: `common.css` + `components.css` + `world-about.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/world-about.html
- ai_instruments/front_3_prepared/world-about.css
- ai_instruments/front_3_prepared/common.css
- ai_instruments/front_3_prepared/components.css
- ai_instruments/front_3_prepared/components/header.html

ДЛЯ РЕФЕРЕНСА (исходный макет):
- ai_instruments/front_3/world_about_page.html
```

**Функциональность**:
- Большая обложка мира с overlay и информацией
- Кнопка "Enter World" для перехода в ленту
- Детальное описание мира в отдельной секции
- История и происхождение мира (timeline)
- Характеристики мира (технологии, магия, социальная структура, география)
- Статистика мира (персонажи, посты, лайки, активность)
- Список избранных персонажей с переходами в профили
- Кнопка "View All Characters"

**Компоненты**: header  
**Особенности**: Двухколоночная сетка, динамическая загрузка персонажей через JavaScript

---

### 7. ✅ Страница входа (`login.html`)

**Исходный макет**: НЕТ (создано с нуля)  
**Готовый файл**: `ai_instruments/front_3_prepared/login.html`  
**Стили**: `common.css` + `auth.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/login.html
- ai_instruments/front_3_prepared/auth.css
- ai_instruments/front_3_prepared/common.css

НЕ ЧИТАТЬ:
- components/header.html (страница без навигации)
- Другие страницы и компоненты
```

**Функциональность**:
- Центрированная форма без header компонента
- Поля: email/username, пароль
- Валидация полей с отображением ошибок
- Checkbox "Remember me"
- Ссылка на восстановление пароля
- Ссылка на регистрацию
- Demo кнопка с предзаполненными данными
- Обработка входа с перенаправлением

**Компоненты**: НЕ использует header.html  
**Особенности**: Автономная страница, JavaScript валидация, demo credentials

---

### 8. ✅ Страница регистрации (`register.html`)

**Исходный макет**: НЕТ (создано с нуля)  
**Готовый файл**: `ai_instruments/front_3_prepared/register.html`  
**Стили**: `common.css` + `auth.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/register.html
- ai_instruments/front_3_prepared/auth.css
- ai_instruments/front_3_prepared/common.css

НЕ ЧИТАТЬ:
- components/header.html (страница без навигации)
- Другие страницы и компоненты
```

**Функциональность**:
- Расширенная форма: username, email, пароль, подтверждение пароля
- Real-time валидация всех полей
- Password requirements индикатор с визуальными подсказками
- Checkbox согласия с Terms of Service и Privacy Policy
- Ссылка на страницу входа
- Обработка регистрации с перенаправлением
- Динамическое включение/отключение кнопки отправки

**Компоненты**: НЕ использует header.html  
**Особенности**: Комплексная валидация, визуальные индикаторы требований пароля

---

### 9. ✅ Профиль персонажа (`character-profile.html`)

**Исходный макет**: `ai_instruments/front_3/character_profile_page.html`  
**Готовый файл**: `ai_instruments/front_3_prepared/character-profile.html`  
**Стили**: `common.css` + `components.css` + `character-profile.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/character-profile.html
- ai_instruments/front_3_prepared/character-profile.css
- ai_instruments/front_3_prepared/common.css
- ai_instruments/front_3_prepared/components.css
- ai_instruments/front_3_prepared/components/header.html

ДЛЯ РЕФЕРЕНСА (исходный макет):
- ai_instruments/front_3/character_profile_page.html
```

**Функциональность**:
- Header персонажа с большим аватаром, именем, ролью, статистикой
- Информация о мире персонажа с переходом
- Детальная биография с чертами характера и специализациями
- Галерея постов персонажа в виде сетки (6 постов)
- Переходы к детальному просмотру постов
- Кнопка "View All Posts"

**Компоненты**: header  
**Особенности**: Динамическая загрузка постов через JavaScript, traits grid

---

### 10. ✅ Настройки пользователя (`settings.html`)

**Исходный макет**: `ai_instruments/front_3/settings_page.html`  
**Готовый файл**: `ai_instruments/front_3_prepared/settings.html`  
**Стили**: `common.css` + `components.css` + `settings.css`

**ЗАВИСИМОСТИ ДЛЯ LLM АГЕНТА**:
```
ОБЯЗАТЕЛЬНО ЧИТАТЬ:
- ai_instruments/front_3_prepared/settings.html
- ai_instruments/front_3_prepared/settings.css
- ai_instruments/front_3_prepared/common.css
- ai_instruments/front_3_prepared/components.css
- ai_instruments/front_3_prepared/components/header.html

ДЛЯ РЕФЕРЕНСА (исходный макет):
- ai_instruments/front_3/settings_page.html
```

**Функциональность**:
- Sidebar навигация с 4 секциями (Account, Credits, Notifications, Privacy)
- Account Settings: редактирование профиля, смена пароля
- Credits & Billing: баланс кредитов, покупка пакетов, история транзакций
- Notifications: toggle switches для различных уведомлений
- Privacy: управление данными, удаление аккаунта
- Кнопки сохранения для каждой секции

**Компоненты**: header  
**Особенности**: Tabbed interface, toggle switches, credit packages, transaction history

---

## 🎯 Инструкции для LLM агента

### При переносе готовых страниц на реальный сайт:

1. **Читать ТОЛЬКО указанные зависимости** для каждой страницы из списка выше
2. **Сохранять структуру и классы** из готовых файлов
3. **Адаптировать данные** под реальные API вызовы
4. **Заменять статические данные** на динамические из бэкенда
5. **Использовать компоненты через загрузку**, как показано в готовых файлах

### Порядок CSS подключения:
```html
<link rel="stylesheet" href="common.css">
<link rel="stylesheet" href="components.css">
<link rel="stylesheet" href="page-specific.css">
```

### Порядок переноса (рекомендуемый):
1. Страницы авторизации (login, register) - без компонентов
2. Главная страница (main) - базовая функциональность
3. Каталог миров (catalog) - списки данных
4. Остальные страницы по приоритету

---

## 🚀 Тестирование

Для тестирования готовых страниц:

```bash
cd ai_instruments/front_3_prepared
python3 server.py
```

**Все доступные страницы**:
- http://localhost:8000/main.html
- http://localhost:8000/catalog.html
- http://localhost:8000/create-world.html
- http://localhost:8000/world-feed.html
- http://localhost:8000/post-detail.html
- http://localhost:8000/world-about.html
- http://localhost:8000/character-profile.html
- http://localhost:8000/settings.html
- http://localhost:8000/login.html
- http://localhost:8000/register.html

---

## 📋 Статус проекта

### ✅ ПОЛНОСТЬЮ ЗАВЕРШЕНО:
- **main.html** + main.css - Главная страница с hero section
- **catalog.html** + catalog.css - Каталог миров с фильтрами  
- **create-world.html** + create-world.css - Создание мира с прогрессом
- **world-feed.html** + feed.css - Лента мира с постами
- **post-detail.html** + дополнения к feed.css - Детальный просмотр с комментариями
- **world-about.html** + world-about.css - Информация о мире
- **login.html** + auth.css - Страница входа
- **register.html** + дополнения к auth.css - Страница регистрации
- **character-profile.html** + character-profile.css - Профиль персонажа
- **settings.html** + settings.css - Настройки пользователя

### 🔄 СТАТУС ПРОЕКТА:
**ВСЕ ЗАДАЧИ ЗАВЕРШЕНЫ** - 10 из 10 страниц готовы ✅  
**Покрытие функциональности**: 100% ✅  
**Статус**: ПОЛНОСТЬЮ ГОТОВ К ПЕРЕНОСУ НА РЕАЛЬНЫЙ САЙТ ✅

---

## 🔗 Навигация между страницами

**Полная схема переходов**:
- main.html → catalog.html, create-world.html
- catalog.html → world-feed.html (через world-card)
- world-feed.html → post-detail.html (через post-card), world-about.html
- world-about.html → world-feed.html, character-profile.html
- post-detail.html → character-profile.html (через аватар), world-feed.html (back)
- create-world.html → world-feed.html (после генерации)
- character-profile.html → world-feed.html, post-detail.html, world-about.html
- settings.html → интеграция с header.html через аватар пользователя
- login.html ↔ register.html

---

**Этот документ является полным техническим заданием для переноса готового фронтенда Generia на реальный сайт.**

**Статус**: ВСЕ СТРАНИЦЫ СОЗДАНЫ И ГОТОВЫ К ПЕРЕНОСУ ✅