# Go Fx Template

Project template for a Go REST API application with Fx DI and built-in web UI

## What's included

- Clean project structure with constructor-based dependency injection
- [Gin](https://github.com/gin-gonic/gin) HTTP server with graceful shutdown with Swagger included
- [GORM](https://gorm.io/) + PostgreSQL storage layer
- [Redis](https://github.com/redis/go-redis) + tokens storage layer
- [MinIO](https://github.com/minio/minio-go) object storage for user avatars and item images
- SMTP email sending for password reset and email verification
- JWT auth with access/refresh tokens, refresh rotation/compromise detection and HttpOnly cookies
- [Viper](https://github.com/spf13/viper) config with YAML file, ENV overrides, validation, and defaults
- Custom logger with console/file output and text/JSON formats
- Embedded static files (templates, styles, assets) for frontend

## Project structure

```text
.
├── cmd/main.go                                # Entrypoint
├── docs/                                      # Generated Swagger files
├── internal/
│   ├── api/
│   │   ├── api.go                             # Gin engine setup, middlewares, routes, Swagger, shutdown
│   │   ├── controllers/
│   │   │   ├── item.go                        # Item API routes and handlers
│   │   │   ├── user.go                        # User/auth API routes and handlers
│   │   │   └── web.go                         # Web UI pages
│   │   ├── cookies/cookies.go                 # Auth cookie helpers
│   │   ├── middlewares/                       # Auth, CORS, request ID
│   │   └── models/                            # Request/response DTOs and API error mapping
│   ├── app/app.go                             # Bootstrap and DI wiring
│   ├── config/config.go                       # Config loading, ENV binding, validation, defaults
│   ├── logger/logger.go                       # Logger setup
│   ├── models/                                # GORM models
│   ├── repository/
│   │   ├── cache/token.go                     # Redis token storage
│   │   └── storage/
│   │       ├── item.go                        # Item storage (Postgres/GORM)
│   │       ├── object.go                      # Object storage (MinIO)
│   │       └── user.go                        # User storage (Postgres/GORM)
│   ├── services/
│   │   ├── auth.go                            # JWT issuing, refresh, revoke, verification
│   │   ├── email.go                           # Email sending and link generation
│   │   ├── image.go                           # Image upload/delete orchestration
│   │   ├── item.go                            # Item business logic
│   │   ├── user.go                            # User business logic
│   │   └── errors/                            # Domain/service errors
│   └── utils/                                 # Small helpers
├── pkg/
│   ├── minio/minio.go                         # MinIO client bootstrap
│   ├── postgres/postgres.go                   # Postgres/GORM bootstrap
│   └── redis/redis.go                         # Redis bootstrap
├── static/
│   ├── embed.go                               # Embedded FS for templates and public assets
│   ├── assets/
│   │   ├── fonts/                             # Fonts
│   │   ├── icons/                             # Icons and favicons
│   │   └── images/                            # Images
│   ├── scripts/                               # Frontend JS
│   ├── styles/                                # Frontend CSS
│   └── templates/                             # Go HTML templates
├── config.yaml                                # Config template
└── go.mod
```

## Config

Config is loaded from `config.yaml` by default. Override the path with `CONFIG_PATH`.

Environment variables take precedence over file values. Full examples and comments are in [`example_config.yaml`](./example_config.yaml).

### Core env vars

| Config key                                | Env variable                  | Default            | Description                                     |
|-------------------------------------------|-------------------------------|--------------------|-------------------------------------------------|
| `app.api.host`                            | `APP_HOST`                    | `0.0.0.0`          | Bind address                                    |
| `app.api.port`                            | `APP_PORT`                    | `8080`             | Listening port                                  |
| `app.api.base_path`                       | `API_BASE_PATH`               | `/api/v1`          | REST API prefix                                 |
| `app.api.gin_release_mode`                | `GIN_RELEASE_MODE`            | `true`             | Hide Gin debug output                           |
| `app.api.shutdown_timeout`                | `WEB_SERVER_SHUTDOWN_TIMEOUT` | `5s`               | Graceful shutdown timeout                       |
| `app.api.web_app.domain`                  | —                             | —                  | Public web app URL used in emails and redirects |
| `app.api.auth.jwt_issuer`                 | `JWT_ISSUER`                  | `go-fx-template`   | JWT issuer                                      |
| `app.api.auth.jwt_audience`               | `JWT_AUDIENCE`                | `go-fx-template`   | JWT audience                                    |
| `app.api.auth.access_token_secret`        | `JWT_ACCESS_TOKEN_SECRET`     | —                  | Access token signing secret                     |
| `app.api.auth.refresh_token_secret`       | `JWT_REFRESH_TOKEN_SECRET`    | —                  | Refresh token signing secret                    |
| `app.api.auth.access_token_ttl`           | `JWT_ACCESS_TOKEN_TTL`        | `24h`              | Access token lifetime                           |
| `app.api.auth.refresh_token_ttl`          | `JWT_REFRESH_TOKEN_TTL`       | `168h`             | Refresh token lifetime                          |
| `app.api.auth.pwd_reset_token_ttl`        | `PWD_RESET_TOKEN_TTL`         | `1h`               | Password reset token lifetime                   |
| `app.api.auth.email_verify_token_ttl`     | `EMAIL_VERIFY_TOKEN_TTL`      | `24h`              | Email verification token lifetime               |
| `app.api.auth.require_email_for_user`     | —                             | `false`            | Require email during registration               |
| `app.api.auth.require_email_verification` | —                             | `false`            | Block login until email is verified             |

### Logger env vars

| Config key                | Env variable       | Default           | Description                                               |
|---------------------------|--------------------|-------------------|-----------------------------------------------------------|
| `app.log.level`           | `LOG_LEVEL`        | `INFO`            | `DEBUG`, `INFO`, `WARN`, `ERROR`                          |
| `app.log.log_format`      | `LOG_FORMAT`       | `text`            | `text` or `json`                                          |
| `app.log.log2console`     | `LOG_TO_CONSOLE`   | `true`            | Log to stdout                                             |
| `app.log.log2file`        | `LOG_TO_FILE`      | `true`            | Log to file                                               |
| `app.log.file_path`       | `LOG_FILE_PATH`    | `application.log` | Log file path                                             |
| `app.log.file_mode`       | `LOG_FILE_MODE`    | `append`          | `append`, `overwrite`, or `rotate`                        |
| `app.log.old_logs_folder` | `LOG_FILES_FOLDER` | –                 | Folder for rotated logs, required when `file_mode=rotate` |

### Database / Redis / MinIO / Email env vars

| Config key                         | Env variable              | Default            | Description                                      |
|------------------------------------|---------------------------|--------------------|--------------------------------------------------|
| `app.database.conn.host`           | `DATABASE_HOST`           | `localhost`        | PostgreSQL host                                  |
| `app.database.conn.port`           | `DATABASE_PORT`           | `5432`             | PostgreSQL port                                  |
| `app.database.conn.user`           | `DATABASE_USER`           | `postgres`         | PostgreSQL user                                  |
| `app.database.conn.password`       | `DATABASE_PASSWORD`       | —                  | PostgreSQL password                              |
| `app.database.conn.database_name`  | `DATABASE_NAME`           | `go-fx-template`   | PostgreSQL database name                         |
| `app.database.conn.ssl_mode`       | `DATABASE_SSL_MODE`       | `disable`          | `disable`, `require`, `verify-ca`, `verify-full` |
| `app.redis.host`                   | `REDIS_HOST`              | `localhost`        | Redis host                                       |
| `app.redis.port`                   | `REDIS_PORT`              | `6379`             | Redis port                                       |
| `app.redis.password`               | `REDIS_PASSWORD`          | —                  | Redis password                                   |
| `app.redis.database`               | `REDIS_DB`                | `0`                | Redis DB index                                   |
| `app.minio.conn.endpoint`          | `MINIO_ENDPOINT`          | `localhost:9000`   | MinIO or S3-compatible endpoint                  |
| `app.minio.conn.access_key_id`     | `MINIO_ACCESS_KEY_ID`     | —                  | Object storage access key                        |
| `app.minio.conn.access_key_secret` | `MINIO_ACCESS_KEY_SECRET` | —                  | Object storage secret key                        |
| `app.minio.conn.use_ssl`           | `MINIO_USE_SSL`           | `false`            | Use HTTPS for object storage                     |
| `app.minio.conn.bucket_name`       | `MINIO_BUCKET_NAME`       | `go-fx-template`   | Bucket name                                      |
| `app.minio.max_upload_size_mb`     | `MINIO_MAX_FILE_SIZE`     | `10`               | Max upload size in MB                            |
| `app.email.host`                   | `EMAIL_HOST`              | `smtp.gmail.com`   | SMTP host                                        |
| `app.email.port`                   | `EMAIL_PORT`              | `587`              | SMTP port                                        |
| `app.email.user`                   | `EMAIL_USER`              | —                  | SMTP username                                    |
| `app.email.password`               | `EMAIL_PASSWORD`          | —                  | SMTP password / app password                     |
| `app.email.from`                   | `EMAIL_FROM`              | —                  | Outgoing `From` header                           |

## Build & Run

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [PostgreSQL]()
- [Redis]()
- [MinIO]() or another S3-compatible object storage

### Local

1. Build
   ```bash
   go build -o go-fx-template ./cmd/main.go
   ```
2. Edit `config.yaml`
   ```bash
   nano config.yaml
   ```
3. Start the required infrastructure services (PostgreSQL, Redis and MinIO)
4. Run
   ```bash
   ./go-fx-template
   ```
   or just
   ```bash
   go run ./cmd
   ```
Web UI will be available on [localhost:8080/](http://localhost:8080/)
Swagger on [localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
API routes on [/api/v1](http://localhost:8080/api/v1)

### Docker

1. Build image
   ```bash
   docker build -t go-fx-template .
   ```
2. Prepare and edit `config.yaml` (or skip and pass ENV later)
   ```bash
   nano config.yaml
   ```
3. Run container
   ```bash
   docker run -d --name go-fx-template \
       -p 8080:8080 \
       -v $(pwd)/config.yaml:/app/config.yaml:ro \
       -v $(pwd)/logs:/app/logs \
       go-fx-template
   ```
