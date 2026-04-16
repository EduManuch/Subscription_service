# Subscription Service

REST API сервис для управления пользовательскими подписками.

## 📌 Функциональность

- Создание подписки
- Получение подписки по ID
- Получение списка подписок с фильтрацией
- Обновление подписки
- Удаление подписки
- Расчет суммарной стоимости подписок за период

---

## 🏗 Архитектура

Проект построен по layered architecture:

```
cmd/
  app/              // точка входа (main.go)

internal/
  config/           // конфигурация приложения
  handlers/         // HTTP слой (REST API)
  logger/           // настройка логирования
  middleware        // 
  model/            // структуры данных
  repository/       // работа с БД
  service/          // бизнес-логика
  storage/          // подключение к БД

migrations/         // SQL миграции
docs/               // swagger документация
```

---

## 🚀 Запуск проекта

### 1. Клонировать репозиторий

```bash
git clone https://github.com/EduManuch/Subscription_service.git
cd Subscription_service
```

---

### 2. Запуск через Docker

```bash
make up 
```

### 3. Применение миграций

```bash
make migrateup
```

---

### 4. Проверка работы

API будет доступен по адресу:

```
http://localhost:8080
```

Swagger UI:

```
http://localhost:8080/swagger/index.html
```

---

## 🗄 База данных

Используется PostgreSQL.

Подключение внутри Docker:

```
host: postgres
port: 5432
db: sub_db
user: postgres
password: password
```

---

## 📡 API эндпоинты

### Создание подписки

```
POST /subscriptions
```

Пример запроса:

```json
{
  "service_name": "Netflix",
  "price": 999,
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "start_date": "01-2025",
  "end_date": "12-2025"
}
```

---

### Получение подписки

```
GET /subscriptions/{id}
```

---

### Список подписок

```
GET /subscriptions?user_id=&service_name=&limit=&offset=
```

---

### Обновление подписки

```
PUT /subscriptions/{id}
```

---

### Удаление подписки

```
DELETE /subscriptions/{id}
```

---

### Расчет суммы

```
GET /subscriptions/total?from=01-2025&to=12-2025
```

---

## 📦 Миграции

Миграции выполняются через `golang-migrate`.

Пример команд:

```bash
make migrateup
make migratedown
```

---

## 🧰 Используемые технологии

- Go
- PostgreSQL
- pgx
- Docker / Docker Compose
- Swagger (swaggo)
- slog (логирование)

---

