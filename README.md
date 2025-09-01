[![Coverage](https://img.shields.io/badge/coverage-88.4%25-brightgreen)](https://github.com/gdyunin/aegis-vault-keeper)
[![Go Version](https://img.shields.io/badge/go-1.24%2B-blue)](https://golang.org/doc/go1.24)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/gdyunin/aegis-vault-keeper/build_up.yml?branch=main&label=build)](https://github.com/gdyunin/aegis-vault-keeper/actions)
[![Docker](https://img.shields.io/badge/docker-ready-blue)](https://hub.docker.com/)

[English version](#aegisvaultkeeper) | [Русская версия](#русская-версия)

# AegisVaultKeeper

AegisVaultKeeper is a secure backend service for storing and managing sensitive personal data: credentials, bank cards, notes, and files. The project features robust data protection, modular architecture, and a modern technology stack for building personal data vaults.

---

## Table of Contents
- [Features](#features)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Key Capabilities and Design Principles](#key-capabilities-and-design-principles)
- [Security](#security)
- [Configuration](#configuration)
- [Makefile Targets](#makefile-targets)
- [Running the Project](#running-the-project)
- [API Documentation](#api-documentation)
- [License](#license)

---

## Features
- Secure storage for:
  - Credentials (login/password)
  - Bank cards
  - Text notes
  - Files and file metadata
- JWT-based authentication
- Data encryption (AES-GCM, bcrypt)
- RESTful API with OpenAPI/Swagger documentation
- Health checks and build info endpoints
- Modular, layered architecture
- Dockerized for local and production use

## Architecture
AegisVaultKeeper follows a modular, layered architecture for maintainability, testability, and security:

- **cmd/server/** — Application entrypoint, initializes DI, config, and starts the HTTP server.
- **internal/server/** — Main business logic and all core modules:
  - **application/** — Application services (use cases) for each domain: auth, bankcard, credential, datasync, filedata, note.
  - **buildinfo/** — Build metadata (version, commit, date) injected at build time.
  - **common/** — Shared utilities and helpers.
  - **config/** — Configuration loading, validation, and extraction (YAML/env).
  - **crypto/** — Cryptographic primitives: AES-GCM, bcrypt, key management.
  - **database/** — PostgreSQL client and DB abstraction.
  - **delivery/** — HTTP delivery layer: routers, middleware, handlers, response formatting, Swagger docs.
    - **about, auth, bankcard, credential, datasync, filedata, health, note/** — HTTP handlers for each domain.
    - **middleware/** — Auth, logging, error handling, request validation.
    - **swagger/** — OpenAPI/Swagger UI integration.
  - **domain/** — Domain models and business rules for each entity (auth, bankcard, credential, filedata, note).
  - **fxshow/** — Dependency injection (Uber Fx) modules and application wiring.
  - **repository/** — Data persistence (PostgreSQL) for each domain, encryption at rest.
  - **security/** — JWT, password hashing, token validation, key generation.
- **pkg/logging/** — Structured logging (zap) for all layers.
- **migrations/** — SQL migration scripts for schema and data.
- **config/** — YAML config files and templates.
- **scripts/** — Utility scripts (e.g., TLS certificate generation).

> Each domain (auth, bankcard, credential, filedata, note) is fully isolated and follows repository-service-delivery layering. All sensitive operations are encapsulated in dedicated modules.

## Technology Stack
- Go 1.24+
- Gin (HTTP framework)
- Uber Fx (dependency injection)
- PostgreSQL (database)
- Docker, Docker Compose
- Swaggo (Swagger/OpenAPI)
- Viper (configuration)
- zap (logging)
- bcrypt, AES-GCM (crypto)

## Key Capabilities and Design Principles
- Secure backend architecture (encryption, JWT, password hashing)
- Clean architecture, separation of concerns
- Dependency injection (Uber Fx)
- REST API design and documentation
- Dockerization and environment configuration
- Automated database migrations
- Modular and testable Go code

## Security
AegisVaultKeeper implements a comprehensive security model:

- **Data Encryption**: All sensitive user data is encrypted at rest using AES-GCM. The master key is provided only via environment variable.
- **Password Hashing**: User passwords are hashed with bcrypt. Plain text passwords are never stored.
- **JWT Authentication**: All API endpoints (except registration/login/health) require JWT tokens signed with a strong HMAC secret.
- **Token Validation Middleware**: Every request with a Bearer token is validated by middleware.
- **TLS**: TLS is supported for all connections. Self-signed certificates are used for development; production requires valid certificates.
- **Config Isolation**: All secrets are injected via environment variables and never committed to version control.
- **Integrity Checks**: File uploads include SHA256 hash calculation for integrity verification.
- **Error Handling**: Authentication and authorization errors are handled with clear, secure error messages and proper HTTP status codes.

> Security is implemented using well-established Go libraries: `crypto/aes`, `crypto/cipher`, `golang.org/x/crypto/bcrypt`, `github.com/golang-jwt/jwt/v5`, and Gin middleware.

## Configuration
All configuration is managed via a combination of environment variables (from `.env`) and YAML file (`config/server.yml`).

- **Environment variables** (set in `.env`): used for secrets and sensitive data (DB password, master key, JWT secret, etc.), as well as most runtime parameters. To create `.env` from a template, run:
  ```bash
  make env-from-template
  # Then edit .env and fill in required secrets and values
  ```
- **YAML file** (`config/server.yml`): used for static and default configuration (non-secret parameters, timeouts, etc.).

> Some parameters (e.g., database credentials, encryption keys) must be set via environment variables for security. Others (timeouts, non-sensitive defaults) are configured in the YAML file. See the table below for details.

### Configuration Fields
| Field / Env Variable         | Description                                      | Example / Values                |
|-----------------------------|--------------------------------------------------|---------------------------------|
| FILE_STORAGE_BASE_PATH      | Base directory for file storage                   | /app/filestorage                |
| POSTGRES_USER               | PostgreSQL username                              | postgres                        |
| POSTGRES_PASSWORD           | PostgreSQL password (required, secret)            | mysecret                        |
| POSTGRES_DB_NAME            | PostgreSQL database name                          | aegisdb                         |
| POSTGRES_HOST               | PostgreSQL host                                  | db                              |
| POSTGRES_PORT               | PostgreSQL port                                  | 5432                            |
| POSTGRES_SSL_MODE           | PostgreSQL SSL mode (disable/require/verify-ca)   | disable                         |
| LOGGER_LEVEL                | Logging level                                    | info, debug, warn, error        |
| TLS_CERT_FILE               | Path to TLS certificate file                      | /app/certs/server.pem           |
| TLS_KEY_FILE                | Path to TLS private key file                      | /app/certs/server-key.pem       |
| MASTER_KEY                  | Master encryption key (required, secret, env var) | (not stored in config file)     |
| ACCESS_TOKEN_LIFETIME       | JWT access token lifetime                         | 24h                             |
| DELIVERY_START_TIMEOUT      | HTTP server start timeout                         | 1s                              |
| DELIVERY_STOP_TIMEOUT       | HTTP server stop timeout                          | 3s                              |
| POSTGRES_INIT_TIMEOUT       | DB init timeout (docker-compose)                  | 31s                             |

> All sensitive values should be set via environment variables and never committed to version control.

## Makefile Targets
AegisVaultKeeper provides a convenient Makefile for common development and CI tasks:

| Target                | Description                                      |
|-----------------------|--------------------------------------------------|
| help                  | List all available make targets                   |
| up                    | Build and start all containers (Docker Compose)   |
| down                  | Stop and remove containers, networks, volumes     |
| restart               | Restart all services                              |
| env-from-template     | Create .env from template if not exists           |
| certs                 | Generate self-signed TLS certificates             |
| deps                  | Update Go dependencies (go mod tidy)              |
| swagdocs              | Generate Swagger/OpenAPI documentation            |
| test                  | Run all tests and show coverage                   |
| lint                  | Run golangci-lint                                 |

> Use `make help` to see all available targets and their descriptions.

## Running the Project

### Quick Start (Docker, production-like)
```bash
# 1. Copy .env from template (if not exists)
make env-from-template
# 2. Edit .env and set all secrets and parameters
# 3. Generate TLS certificates for HTTPS (local only)
make certs
# 4. Build and start all services (PostgreSQL, migrations, backend)
make up
# 5. API will be available at https://localhost:56789/api
```

### Local Development (Go)
```bash
# 1. Copy .env from template and configure variables
make env-from-template
# 2. Generate TLS certificates
make certs
# 3. Start PostgreSQL via docker-compose (or manually)
docker-compose up -d db
# 4. Apply migrations (via docker-compose or manually)
docker-compose up migrate
# 5. Install dependencies and generate Swagger docs
make deps
make swagdocs
# 6. Run the server locally
go run ./cmd/server
```

- To stop and clean up: `make down`
- To run tests: `make test`
- To lint: `make lint`

## API Documentation
- Swagger UI: [https://localhost:56789/swagger/index.html](https://localhost:56789/swagger/index.html)
- OpenAPI spec: `docs/swagger.yaml`

## License
MIT

---

# Русская версия

AegisVaultKeeper — это backend-сервис для безопасного хранения и управления персональными данными: учетные данные, банковские карты, заметки, файлы. Проект обеспечивает надёжную защиту данных, модульную архитектуру и современный технологический стек для построения защищённых хранилищ персональной информации.

---

## Оглавление
- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Технологический стек](#технологический-стек)
- [Ключевые возможности и принципы проектирования](#ключевые-возможности-и-принципы-проектирования)
- [Безопасность](#безопасность)
- [Конфигурация](#конфигурация)
- [Цели Makefile](#цели-makefile)
- [Запуск проекта](#запуск-проекта)
- [Документация API](#документация-api)
- [Лицензия](#лицензия)

---

## Возможности
- Безопасное хранение:
  - Учетные данные (логин/пароль)
  - Банковские карты
  - Текстовые заметки
  - Файлы и метаданные
- Аутентификация через JWT
- Шифрование данных (AES-GCM, bcrypt)
- RESTful API с документацией OpenAPI/Swagger
- Эндпоинты для проверки статуса и информации о сборке
- Модульная архитектура
- Docker-окружение для локальной и продакшн-среды

## Архитектура
AegisVaultKeeper следует модульной и многослойной архитектуре для обеспечения удобства сопровождения, тестирования и безопасности:

- **cmd/server/** — Точка входа в приложение, инициализирует DI, конфигурирует и запускает HTTP-сервер.
- **internal/server/** — Основная бизнес-логика и все ключевые модули:
  - **application/** — Сервисы приложений (варианты использования) для каждого домена: auth, bankcard, credential, datasync, filedata, note.
  - **buildinfo/** — Метаданные сборки (версия, коммит, дата), внедряемые во время сборки.
  - **common/** — Общие утилиты и вспомогательные функции.
  - **config/** — Загрузка, валидация и извлечение конфигурации (YAML/переменные окружения).
  - **crypto/** — Криптографические примитивы: AES-GCM, bcrypt, управление ключами.
  - **database/** — Клиент PostgreSQL и абстракция БД.
  - **delivery/** — Уровень доставки HTTP: маршрутизаторы, промежуточное ПО, обработчики, форматирование ответов, документация Swagger.
    - **about, auth, bankcard, credential, datasync, filedata, health, note/** — Обработчики HTTP для каждого домена.
    - **middleware/** — Аутентификация, логирование, обработка ошибок, валидация запросов.
    - **swagger/** — Интеграция OpenAPI/Swagger UI.
  - **domain/** — Модели домена и бизнес-правила для каждой сущности (auth, bankcard, credential, filedata, note).
  - **fxshow/** — Внедрение зависимостей (Uber Fx) и связывание приложений.
  - **repository/** — Сохранение данных (PostgreSQL) для каждого домена, шифрование на диске.
  - **security/** — JWT, хеширование паролей, валидация токенов, генерация ключей.
- **pkg/logging/** — Структурированное логирование (zap) для всех слоев.
- **migrations/** — SQL-скрипты миграции для схемы и данных.
- **config/** — YAML-файлы конфигурации и шаблоны.
- **scripts/** — Утилиты (например, генерация TLS-сертификатов).

> Каждый домен (auth, bankcard, credential, filedata, note) полностью изолирован и следует слоистой архитектуре repository-service-delivery. Все чувствительные операции инкапсулированы в специализированные модули.

## Технологический стек
- Go 1.24+
- Gin (HTTP-фреймворк)
- Uber Fx (внедрение зависимостей)
- PostgreSQL (БД)
- Docker, Docker Compose
- Swaggo (Swagger/OpenAPI)
- Viper (конфигурирование)
- zap (логирование)
- bcrypt, AES-GCM (криптография)

## Ключевые возможности и принципы проектирования
- Безопасная архитектура backend (шифрование, JWT, хеширование паролей)
- Чистая архитектура, разделение ответственности
- Внедрение зависимостей (Uber Fx)
- Проектирование и документирование REST API
- Dockerизация и настройка окружения
- Автоматизация миграций БД
- Модульный и тестируемый Go-код

## Безопасность
AegisVaultKeeper реализует комплексную модель безопасности:

- **Шифрование данных**: Все чувствительные пользовательские данные шифруются на диске с помощью AES-GCM. Мастер-ключ задается только через переменную окружения.
- **Хеширование паролей**: Пароли пользователей хешируются с помощью bcrypt. Пароли никогда не сохраняются в открытом виде.
- **Аутентификация JWT**: Все API-эндпоинты (кроме регистрации/логина/health) требуют JWT-токен, подписанный HMAC-секретом.
- **Промежуточная проверка токена**: Каждый запрос с Bearer-токеном проходит проверку в middleware.
- **TLS**: Сервер поддерживает TLS для всех соединений. Для разработки используются самоподписанные сертификаты; для продакшена требуются валидные сертификаты.
- **Изоляция конфигурации**: Все секреты передаются только через переменные окружения и не попадают в систему контроля версий.
- **Проверка целостности**: При загрузке файлов вычисляется SHA256-хеш для проверки целостности.
- **Обработка ошибок**: Ошибки аутентификации и авторизации обрабатываются с понятными и безопасными сообщениями и корректными HTTP-статусами.

> Все механизмы безопасности реализованы с использованием проверенных Go-библиотек: `crypto/aes`, `crypto/cipher`, `golang.org/x/crypto/bcrypt`, `github.com/golang-jwt/jwt/v5` и middleware Gin.

## Конфигурация
Вся конфигурация управляется через переменные окружения и YAML (`config/server.yml`).

- **Переменные окружения** (устанавливаются в `.env`): используются для секретов и чувствительных данных (пароль БД, мастер-ключ, секрет JWT и т.д.), а также для большинства параметров выполнения. Чтобы создать `.env` из шаблона, выполните:
  ```bash
  make env-from-template
  # Затем отредактируйте .env и заполните необходимые секреты и значения
  ```
- **YAML-файл** (`config/server.yml`): используется для статической и стандартной конфигурации (неконфиденциальные параметры, таймауты и т.д.).

> Некоторые параметры (например, учетные данные базы данных, ключи шифрования) должны быть установлены через переменные окружения по соображениям безопасности. Другие (таймауты, не чувствительные по умолчанию) настраиваются в YAML-файле. См. таблицу ниже для получения дополнительной информации.

### Описание параметров конфигурации
| Поле / Переменная окружения  | Описание                                         | Пример / Значения               |
|-----------------------------|--------------------------------------------------|---------------------------------|
| FILE_STORAGE_BASE_PATH      | Базовая директория для хранения файлов            | /app/filestorage                |
| POSTGRES_USER               | Имя пользователя PostgreSQL                       | postgres                        |
| POSTGRES_PASSWORD           | Пароль PostgreSQL (обязательно, секретно)         | mysecret                        |
| POSTGRES_DB_NAME            | Имя базы данных PostgreSQL                        | aegisdb                         |
| POSTGRES_HOST               | Хост PostgreSQL                                  | db                              |
| POSTGRES_PORT               | Порт PostgreSQL                                  | 5432                            |
| POSTGRES_SSL_MODE           | Режим SSL для PostgreSQL (disable/require/verify-ca) | disable                     |
| LOGGER_LEVEL                | Уровень логирования                              | info, debug, warn, error        |
| TLS_CERT_FILE               | Путь к TLS-сертификату                           | /app/certs/server.pem           |
| TLS_KEY_FILE                | Путь к приватному TLS-ключу                      | /app/certs/server-key.pem       |
| MASTER_KEY                  | Мастер-ключ шифрования (обязательно, секретно, env) | (не хранится в файле конфига) |
| ACCESS_TOKEN_LIFETIME       | Время жизни JWT access token                      | 24h                             |
| DELIVERY_START_TIMEOUT      | Таймаут запуска HTTP-сервера                      | 1s                              |
| DELIVERY_STOP_TIMEOUT       | Таймаут остановки HTTP-сервера                    | 3s                              |
| POSTGRES_INIT_TIMEOUT       | Таймаут инициализации БД (docker-compose)         | 31s                             |

> Все чувствительные значения должны задаваться только через переменные окружения и не попадать в систему контроля версий.

## Цели Makefile
AegisVaultKeeper предоставляет удобный Makefile для основных задач разработки и CI:

| Цель                  | Описание                                         |
|-----------------------|--------------------------------------------------|
| help                  | Показать все доступные цели make                  |
| up                    | Собрать и запустить все контейнеры (Docker Compose) |
| down                  | Остановить и удалить контейнеры, сети, тома       |
| restart               | Перезапустить все сервисы                        |
| env-from-template     | Создать .env из шаблона, если не существует       |
| certs                 | Сгенерировать самоподписанные TLS-сертификаты     |
| deps                  | Обновить зависимости Go (go mod tidy)             |
| swagdocs              | Сгенерировать документацию Swagger/OpenAPI        |
| test                  | Запустить все тесты и показать покрытие           |
| lint                  | Запустить golangci-lint                           |

> Используйте `make help` для просмотра всех целей и их описаний.

## Запуск проекта

### Быстрый старт (Docker, production-like)
```bash
# 1. Скопируйте .env из шаблона (если файла нет)
make env-from-template
# 2. Отредактируйте .env и укажите все секреты и параметры
# 3. Сгенерируйте TLS-сертификаты для HTTPS (локально)
make certs
# 4. Соберите и запустите все сервисы (PostgreSQL, миграции, backend)
make up
# 5. API будет доступен на https://localhost:56789/api
```

### Локальная разработка (Go)
```bash
# 1. Скопируйте .env из шаблона и настройте переменные
make env-from-template
# 2. Сгенерируйте TLS-сертификаты
make certs
# 3. Запустите PostgreSQL через docker-compose (или вручную)
docker-compose up -d db
# 4. Примените миграции (docker-compose или вручную)
docker-compose up migrate
# 5. Установите зависимости и сгенерируйте Swagger-документацию
make deps
make swagdocs
# 6. Запустите сервер локально
go run ./cmd/server
```

- Для остановки и очистки окружения: `make down`
- Для тестирования: `make test`
- Для линтинга: `make lint`

## Документация API
- Swagger UI: [https://localhost:56789/swagger/index.html](https://localhost:56789/swagger/index.html)
- OpenAPI: `docs/swagger.yaml`

## Лицензия
MIT