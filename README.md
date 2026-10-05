# User Service
Backend-сервис для регистрации и аутентификации пользователей. Проект написан на Go с использованием стандартного HTTP-сервера и PostgreSQL.

## Возможности
Сервис поддерживает регистрацию пользователей, вход по логину и паролю, создание JWT access token, проверку JWT и разграничение доступа по пользователю и роли.
Поддерживаемые роли:
* `user` — обычный пользователь
* `admin` — администратор
Обычный пользователь может изменять и удалять только собственный аккаунт. Администратор может изменять и удалять любого пользователя.

## API
### Регистрация

```http
POST /api/v1/register
Content-Type: application/json
```

Тело запроса:

```json
{
  "name": "Kirill",
  "login": "kirill",
  "password": "mypassword123"
}
```
При успешной регистрации создаётся пользователь с ролью `user` и возвращается JWT.

Ответ:

```http
201 Created
```

### Вход

```http
POST /api/v1/login
Content-Type: application/json
```

Тело запроса:

```json
{
  "login": "kirill",
  "password": "mypassword123"
}
```

При успешной аутентификации сервис возвращает JWT access token.

Для доступа к защищённым endpoint'ам токен передаётся в заголовке:

```http
Authorization: Bearer <JWT>
```

### Получение пользователя

```http
GET /api/v1/user/{id}
Authorization: Bearer <JWT>
```

### Изменение пользователя

```http
PUT /api/v1/user/{id}
Authorization: Bearer <JWT>
Content-Type: application/json
```

Тело:

```json
{
  "name": "New Name",
  "password": "newpassword123"
}
```

Обычный пользователь может изменить только собственный профиль. Администратор может изменить профиль любого пользователя.

### Удаление пользователя

```http
DELETE /api/v1/user/{id}
Authorization: Bearer <JWT>
```

Обычный пользователь может удалить только собственный аккаунт. Администратор может удалить любого пользователя.

## Authentication

Для хранения паролей используется bcrypt. В PostgreSQL сохраняется не исходный пароль, а его bcrypt-хеш.

Во время входа происходит следующая последовательность:

```text
login + password
        ↓
поиск пользователя в PostgreSQL
        ↓
bcrypt.CompareHashAndPassword
        ↓
JWT
```

JWT содержит идентификатор пользователя и его роль.

Упрощённо payload токена выглядит следующим образом:

```json
{
  "user_id": 15,
  "user_role": "user",
  "iat": 1780000000,
  "exp": 1780000900
}
```

JWT подписывается с использованием HS256 и секретного ключа, который загружается из переменной окружения.

Access token имеет ограниченное время жизни.

## Authorization

После проверки JWT middleware помещает информацию об аутентифицированном пользователе в context текущего HTTP-запроса:

```text
AuthUser
├── ID
└── Role
```

Далее конкретный handler проверяет права пользователя.

Правила текущей версии:

```text
user
 ├── update own user
 └── delete own user

admin
 ├── update any user
 └── delete any user
```

Authentication и authorization разделены:

```text
Authentication
    ↓
Кто пользователь?

Authorization
    ↓
Что этому пользователю разрешено?
```

## Архитектура

Проект разделён на несколько слоёв:

```text
HTTP
 ↓
Handler
 ↓
Service
 ↓
Repository
 ↓
PostgreSQL
```

### Handler

Отвечает за HTTP:

* маршрутизацию
* чтение JSON
* HTTP status codes
* формирование JSON-ответов
* authentication middleware
* проверку прав доступа

### Service

Содержит бизнес-логику:

* валидацию данных
* регистрацию пользователя
* bcrypt
* проверку credentials
* преобразование ошибок repository в ошибки service

### Repository

Отвечает только за работу с PostgreSQL:

* INSERT
* SELECT
* UPDATE
* DELETE

### Database

Содержит создание и проверку PostgreSQL connection pool на основе `pgxpool`.

## Структура проекта

```text
internal/
├── config/
├── database/
├── handler/
│   ├── handler.go
│   └── user.go
├── model/
│   └── user.go
├── repository/
│   └── user.go
├── server/
│   └── server.go
└── service/
    ├── auth.go
    └── user.go
```

## Конфигурация

Сервис получает настройки из переменных окружения.

Пример:

```text
DatabaseDSN=postgres://postgres:password@localhost:5432/taskflow?sslmode=disable
JWTSecretkey=your-secret-key
Port=8080
Address=localhost
```

`DatabaseDSN` и `JWTSecretkey` являются обязательными.

Пример запуска:

```bash
go run .
```

## PostgreSQL

Таблица пользователей должна содержать как минимум следующие поля:

```text
id
name
login
password
role
```

`login` должен быть уникальным.

`role` должна принимать только поддерживаемые значения:

```text
user
admin
```

Пароль хранится в виде bcrypt-хеша.

## Обработка HTTP-ошибок

Используются стандартные HTTP-коды:

```text
400 Bad Request
```

Некорректный JSON или входные данные.

```text
401 Unauthorized
```

Отсутствующий или недействительный JWT.

```text
403 Forbidden
```

JWT валиден, но у пользователя недостаточно прав.

```text
404 Not Found
```

Пользователь не существует.

```text
500 Internal Server Error
```

Внутренняя ошибка приложения или базы данных.

## Graceful Shutdown

HTTP-сервер использует `http.Server` с timeout:

```text
ReadTimeout  = 5s
WriteTimeout = 10s
IdleTimeout  = 60s
```

При `SIGINT` или `SIGTERM` запускается graceful shutdown с ограничением в 5 секунд.
