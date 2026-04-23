# 🚀 Catalyst

**Enterprise-Grade Event-Driven Commerce Backend**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-1e90ff?style=for-the-badge)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
[![Event-Driven](https://img.shields.io/badge/Pattern-Event%20Driven-ff6b6b?style=for-the-badge)](https://en.wikipedia.org/wiki/Event-driven_architecture)

### What is Catalyst?

Catalyst demonstrates how to build a **production-grade e-commerce platform** 
using Go and microservices patterns. The project focuses on:

- 🏗️ **Clean Architecture** — Clear separation of concerns for maintainability
- ⚡ **Event-Driven Design** — Kafka-based async processing without data loss
- 🔐 **Transaction Safety** — Atomic operations, server-side validation
- 📊 **Real-time Performance** — Redis caching, singleflight cache stampede prevention
- 🎯 **Production Ready** — Graceful shutdown, structured logging, error handling

### Why "Catalyst"?

In chemistry, a catalyst **accelerates a reaction without being consumed**. 
Similarly, Catalyst is the infrastructure layer that makes your commerce 
system work reliably — transparent, fast, and essential. It's not the 
business logic; it's what makes good business logic possible.

---

## ✨ Fitur Utama

<table>
<tr>
<td width="50%">

### 👤 User Management

- ✅ User registration & authentication
- ✅ JWT-based authorization
- ✅ Password hashing with bcrypt
- ✅ Profile management (CRUD)

</td>
<td width="50%">

### 📦 Product Management

- ✅ CRUD operations for products
- ✅ Category management
- ✅ Redis caching with auto-invalidation
- ✅ Cache stampede protection (singleflight)
- ✅ Pagination support

</td>
</tr>
<tr>
<td width="50%">

### 💳 Transaction Management

- ✅ Order creation & tracking
- ✅ Transaction history
- ✅ UUID-based transaction IDs
- ✅ Business logic validation
- ✅ Event publishing to Kafka

</td>
<td width="50%">

### 🛠️ Technical Features

- ✅ Auto database migration
- ✅ Input validation (struct-level)
- ✅ Error handling middleware
- ✅ Dependency Injection (Wire)
- ✅ Message broker integration

</td>
</tr>
</table>

---

## 🎯 Why This Project Matters

This project is not just another e-commerce API. It's a **reference implementation** 
demonstrating how to build reliable, scalable systems with Go. Each design decision 
solves real production problems:

### Problem: Data Integrity
**Scenario:** Two requests charge customer simultaneously → double charge disaster

**Solution:** Atomic database transactions. All changes succeed together or 
all rollback. No partial states.

### Problem: Performance Under Load
**Scenario:** 1000 concurrent requests for "iPhone 15" → database melts

**Solution:** Redis caching + singleflight. Only 1 database query, 999 requests 
get cached result instantly.

### Problem: System Reliability
**Scenario:** Email service is down → should transaction fail?

**Solution:** Event-driven architecture. Transaction completes, then async 
Kafka consumer handles email. If email fails, retry later. Transaction safe either way.

### Problem: Maintainability at Scale
**Scenario:** Switch database from MySQL to PostgreSQL → rewrite everything?

**Solution:** Clean architecture. Repository layer abstracts database. 
Only repository changes, service/handler untouched.

---

## 🚀 Use Cases

Catalyst is designed for:

- **E-commerce platforms** needing reliable order processing
- **Fintech applications** requiring atomic transactions
- **Marketplace systems** with inventory management
- **Subscription services** with event-based workflows

Learn how Catalyst handles these with pattern that scale to millions of users.

---

## 🏗️ Arsitektur

Aplikasi ini menggunakan **Clean Architecture** yang terintegrasi dengan **Event-Driven Architecture** untuk menopang platform *commerce* yang sangat *scalable* dan tangguh (*resilient*).

### Layered Architecture Diagram

<div align="center">

```text
       [Client Request / HTTP]
                 │
                 ▼
 ┌───────────────────────────────┐
 │       HTTP Handler (Delivery) │  ← Menerima input, parsing DTO, HTTP Response
 └───────────────┬───────────────┘
                 │
                 ▼
 ┌───────────────────────────────┐  ← Validasi bisnis, kalkulasi harga, orkestrasi
 │      Service (Use Case)       │  
 │   [ Transaction Atomicity ]   │─────┐ (Publish Async Event)
 └───────────────┬───────────────┘     │
                 │                     ▼
                 ▼              ┌────────────────┐
 ┌───────────────────────────────┐      │ Apache Kafka   │  ← Event Broker untuk Asynchronous Task
 │   Repository (Data Access)    │      │ (Message Bus)  │  
 │   [ Cache Stampede Protect ]  │      └────────────────┘
 └───────────────┬───────────────┘             │
                 │                             ▼
                 │                      ┌────────────────┐
         ┌───────┴───────┐              │  Async Workers │  ← Notifikasi (Email), Third-party integrations
         ▼               ▼              └────────────────┘
 ┌──────────────┐ ┌──────────────┐
 │    Redis     │ │    MySQL     │
 │ (Multi-tier  │ │ (Source of   │
 │   Caching)   │ │    Truth)    │
 └──────────────┘ └──────────────┘
```

</div>

**Setiap Layer memiliki batasan yang tegas:**
- **Handler (Presentation)**: Hanya fokus pada HTTP layer, deserialization JSON -> DTO, dan pemetaan *Error Sentinel* menjadi HTTP Status Code yang tepat.
- **Service (Business Logic)**: Tidak tahu soal *database* spesifik atau HTTP. Menjalankan *Use Cases* utama (misal: memotong stok, memastikan harga dari DB, melempar Event).
- **Repository (Data Access)**: Berfokus mengeksekusi query database dan caching (Redis). Di sinilah `Singleflight` digunakan untuk mencegah *Cache Stampede*.
- **Infrastructure**: Inisialisasi dependensi eskternal (Koneksi Database, Redis Client, Kafka Writer).

### 📂 Struktur Folder

```
go-restfull-api/
├── cmd/
│   ├── api/                 # Main application entrypoint
│   └── injector/            # Dependency injection (Wire)
├── config/                  # Configuration management (Viper)
├── db/migrations/           # Database migrations (golang-migrate)
├── docs/                    # Documentation & Swagger specs
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/     # HTTP handlers (Presentation layer)
│   │       ├── dto/         # Data Transfer Objects
│   │       ├── middleware/  # JWT Auth, Error handling
│   │       └── route/       # Route definitions
│   ├── domain/
│   │   ├── *.go            # Business entities (User, Product, Transaction)
│   │   └── event/          # Event payloads (OrderCreatedEvent)
│   ├── infrastructure/      # External service clients
│   │   ├── database.go     # MySQL connection
│   │   ├── redis.go        # Redis connection
│   │   └── kafka.go        # Kafka writer setup
│   ├── pkg/
│   │   ├── cache/          # Cache wrapper (abstraction)
│   │   └── response/       # Standardized HTTP responses
│   ├── producer/           # Kafka producers (OrderProducer)
│   ├── repository/
│   │   └── mysql/          # Data access layer (MySQL queries)
│   └── service/            # Business logic layer
└── test/                    # Integration & unit tests
```

---

## 🛠️ Tech Stack

<div align="center">
<table>
<tr>
<td align="center" width="20%">
<img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" width="80" height="80" alt="Go"/>
<br><b>Go 1.22+</b>
<br>Core Language
</td>
<td align="center" width="20%">
<img src="https://www.mysql.com/common/logos/logo-mysql-170x115.png" width="80" height="80" alt="MySQL"/>
<br><b>MySQL 8.0</b>
<br>Primary Database
</td>
<td align="center" width="20%">
<img src="https://redis.io/wp-content/uploads/2024/04/Logotype.svg?auto=webp&quality=85,75&width=120" width="80" alt="Redis"/>
<br><b>Redis 7.0</b>
<br>Caching Layer
</td>
<td align="center" width="20%">
<img src="https://img.icons8.com/?size=100&id=k4fZIepXxmAZ&format=png&color=ffffff" width="80" alt="Kafka"/>
<br><b>Apache Kafka</b>
<br>Message Broker
</td>
<td align="center" width="20%">
<img src="https://www.docker.com/wp-content/uploads/2022/03/vertical-logo-monochromatic.png" width="80" height="80" alt="Docker"/>
<br><b>Docker</b>
<br>Containerization
</td>
</tr>
</table>
</div>

### 📚 Dependencies & Libraries

| Category           | Library                                                                   | Purpose                           |
| ------------------ | ------------------------------------------------------------------------- | --------------------------------- |
| **Router**         | [`julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) | High-performance HTTP router      |
| **Database**       | [`go-sql-driver/mysql`](https://github.com/go-sql-driver/mysql)           | MySQL driver for Go               |
| **Cache**          | [`redis/go-redis`](https://github.com/redis/go-redis)                     | Redis client for Go               |
| **Message Broker** | [`segmentio/kafka-go`](https://github.com/segmentio/kafka-go)             | Pure Go Kafka client              |
| **Concurrency**    | [`golang.org/x/sync`](https://pkg.go.dev/golang.org/x/sync)               | Singleflight (cache stampede)     |
| **Validation**     | [`go-playground/validator`](https://github.com/go-playground/validator)   | Struct validation                 |
| **JWT**            | [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt)                     | JSON Web Token implementation     |
| **Config**         | [`spf13/viper`](https://github.com/spf13/viper)                           | Configuration management          |
| **DI**             | [`google/wire`](https://github.com/google/wire)                           | Compile-time dependency injection |
| **Migration**      | [`golang-migrate`](https://github.com/golang-migrate/migrate)             | Database migrations               |
| **Password**       | [`golang.org/x/crypto`](https://pkg.go.dev/golang.org/x/crypto)           | Bcrypt hashing                    |
| **UUID**           | [`google/uuid`](https://github.com/google/uuid)                           | UUID generation                   |
| **Testing**        | [`stretchr/testify`](https://github.com/stretchr/testify)                 | Assertions & Mocking framework    |
| **Testing Mocks**  | [`DATA-DOG/go-sqlmock`](https://github.com/DATA-DOG/go-sqlmock)           | Mocking SQL driver behaviour      |

---

## 🚀 Quick Start

### 📋 Prerequisites

Pastikan sistem Anda telah menginstall:

- [Git](https://git-scm.com/) (v2.0+)
- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

### ⚙️ Installation

#### 1️⃣ Clone Repository

```bash
git clone https://github.com/assidik12/go-restfull-api.git
cd go-restfull-api
```

#### 2️⃣ Setup Environment Variables

Buat file `.env` di root directory:

```bash
# Windows (CMD)
type nul > .env

# Windows (PowerShell)
New-Item .env -ItemType File

# Linux/Mac
touch .env
```

Copy dan sesuaikan konfigurasi berikut ke file `.env`:

```env
# ================================
# Application Configuration
# ================================
APP_PORT=3000

# ================================
# MySQL Database Configuration
# ================================
MYSQL_HOST=db
MYSQL_PORT=3306
MYSQL_USER=gouser
MYSQL_PASSWORD=gosecret123
MYSQL_DATABASE=go_ecommerce_db
MYSQL_ROOT_PASSWORD=rootsecret123

# Database URL for migrations
DB_URL=mysql://gouser:gosecret123@tcp(db:3306)/go_ecommerce_db?multiStatements=true

# ================================
# Redis Cache Configuration
# ================================
REDIS_HOST=cache
REDIS_PORT=6379
REDIS_PASSWORD=redissecret123

# ================================
# Kafka Configuration
# ================================
KAFKA_BROKER=message-broker:9092
KAFKA_HOST=message-broker
KAFKA_PORT=9092

# ================================
# JWT Configuration
# ================================
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

> ⚠️ **Security Warning**:
>
> - Ganti semua password dengan nilai yang strong untuk production
> - Pastikan `.env` sudah ada di `.gitignore`
> - Jangan commit `.env` ke repository

#### 3️⃣ Run Application

```bash
docker-compose up --build
```

Proses ini akan:

- 📦 Build Docker images untuk Go application
- 🗄️ Setup MySQL database dengan healthcheck
- 🚀 Setup Redis cache dengan healthcheck
- 📨 Setup Apache Kafka & Zookeeper
- 🔄 Menjalankan database migrations secara otomatis
- ▶️ Start aplikasi pada port 3001

### ✅ Verifikasi

Setelah semua container berjalan, Anda akan melihat output:

```
✅ zookeeper              - healthy
✅ kafka                  - healthy
✅ db-mysql-service       - healthy
✅ redis-cache-service    - healthy
✅ go-app-service         - running
```

Akses endpoints:

- 🌐 **API Base URL**: http://localhost:3001
- 📚 **API Documentation**: http://localhost:3001/api/v1/docs
- 📊 **Kafka Broker**: `localhost:9092`
- 🗄️ **MySQL**: `localhost:3307`
- 💾 **Redis**: `localhost:6379`

---

## 📊 Services Overview

### 🐳 Docker Services

| Service            | Container Name        | Image                    | Port(s)                  | Volume       | Description         |
| ------------------ | --------------------- | ------------------------ | ------------------------ | ------------ | ------------------- |
| **go-app-service** | `go-app-service`      | Custom (built)           | `3001:3000`              | -            | Main Go application |
| **db**             | `db-mysql-service`    | `mysql:8.0`              | `3307:3306`              | `db-data`    | MySQL database      |
| **cache**          | `redis-cache-service` | `redis:7.0-alpine`       | `6379:6379`              | `redis-data` | Redis cache         |
| **zookeeper**      | `zookeeper`           | `wurstmeister/zookeeper` | `2181:2181`              | -            | Kafka coordination  |
| **kafka**          | `kafka`               | `wurstmeister/kafka`     | `9092:9092`, `9093:9093` | `kafka-data` | Message broker      |

### 🔌 Port Mapping

| Service   | Internal Port | External Port | Access URL              | Description        |
| --------- | ------------- | ------------- | ----------------------- | ------------------ |
| Go API    | 3000          | 3001          | `http://localhost:3001` | HTTP REST API      |
| MySQL     | 3306          | 3307          | `localhost:3307`        | Database client    |
| Redis     | 6379          | 6379          | `localhost:6379`        | Cache client       |
| Kafka     | 9092          | 9092          | `localhost:9092`        | Kafka broker       |
| Zookeeper | 2181          | 2181          | `localhost:2181`        | Kafka coordination |

### 💾 Data Persistence

- **MySQL Data**: Persisted in Docker volume `db-data`
- **Redis Data**: Persisted in Docker volume `redis-data`
- **Kafka Data**: Persisted in Docker volume `kafka-data`
- **Migrations**: Auto-run on container startup via `entrypoint.sh`

---

## 🔧 Configuration

### 🌍 Environment Variables

<details>
<summary><b>Click to expand full configuration reference</b></summary>

#### Application Settings

| Variable   | Default | Description                       |
| ---------- | ------- | --------------------------------- |
| `APP_PORT` | `3000`  | Port untuk aplikasi Go (internal) |

#### MySQL Settings

| Variable              | Required | Description                               |
| --------------------- | -------- | ----------------------------------------- |
| `MYSQL_HOST`          | ✅       | Database host (gunakan `db` untuk Docker) |
| `MYSQL_PORT`          | ✅       | Database port (default: `3306`)           |
| `MYSQL_USER`          | ✅       | Database username                         |
| `MYSQL_PASSWORD`      | ✅       | Database password                         |
| `MYSQL_DATABASE`      | ✅       | Database name                             |
| `MYSQL_ROOT_PASSWORD` | ✅       | MySQL root password                       |
| `DB_URL`              | ✅       | Full connection string untuk migrations   |

#### Redis Settings

| Variable         | Required | Description                               |
| ---------------- | -------- | ----------------------------------------- |
| `REDIS_HOST`     | ✅       | Redis host (gunakan `cache` untuk Docker) |
| `REDIS_PORT`     | ✅       | Redis port (default: `6379`)              |
| `REDIS_PASSWORD` | ✅       | Redis authentication password             |

#### Kafka Settings

| Variable       | Required | Description                                |
| -------------- | -------- | ------------------------------------------ |
| `KAFKA_BROKER` | ✅       | Kafka broker address (format: `host:port`) |

#### Security Settings

| Variable     | Required | Description                           |
| ------------ | -------- | ------------------------------------- |
| `JWT_SECRET` | ✅       | Secret key untuk JWT token generation |

</details>

---

## 📚 Dokumentasi API

### 📖 Swagger Documentation

API documentation tersedia melalui Swagger UI:

**URL**: http://localhost:3001/api/v1/docs/

### 🔑 Authentication

API menggunakan **JWT (JSON Web Token)** untuk authentication:

1. Register user melalui endpoint `/api/v1/users/register`
2. Login untuk mendapatkan JWT token via `/api/v1/users/login`
3. Include token di header: `Authorization: Bearer <your-token>`

### 📍 Endpoints Overview

<details>
<summary><b>Click to see available endpoints</b></summary>

#### User Endpoints

- `POST /api/v1/users/register` - Register new user
- `POST /api/v1/users/login` - Login user (returns JWT)
- `GET /api/v1/users/profile` - Get user profile (🔒 protected)
- `PUT /api/v1/users/profile` - Update user profile (🔒 protected)

#### Product Endpoints

- `GET /api/v1/products` - Get all products with pagination (cached ⚡)
- `GET /api/v1/products/:id` - Get product by ID (cached ⚡)
- `POST /api/v1/products` - Create new product (🔒 protected)
- `PUT /api/v1/products/:id` - Update product (🔒 protected, invalidates cache)
- `DELETE /api/v1/products/:id` - Delete product (🔒 protected, invalidates cache)

#### Transaction Endpoints

- `GET /api/v1/transactions` - Get all user transactions (🔒 protected)
- `GET /api/v1/transactions/:id` - Get transaction by ID (🔒 protected)
- `POST /api/v1/transactions` - Create transaction (🔒 protected, publishes event 📨)

</details>

---

## 🎯 Caching Strategy

### Redis Implementation

Aplikasi ini menggunakan **Redis** untuk caching data produk guna mengurangi beban database dan meningkatkan response time.

#### Cache Specifications

- **Cached Endpoints**:
  - `GET /api/v1/products/:id` - Detail produk individual
  - `GET /api/v1/products?page=X` - Daftar produk dengan paginasi
- **TTL (Time-To-Live)**: 10 menit
- **Cache Key Pattern**:
  - Detail: `product:detail:{id}`
  - List: `products:list:page:{page_number}`
- **Strategy**: Cache-Aside (Lazy Loading)

#### Cache Flow

```
┌─────────────────┐
│  Client Request │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│  Check Redis    │─────▶│  Cache HIT   │──┐
│     Cache       │      └──────────────┘  │
└────────┬────────┘                        │
         │ Cache MISS                      │
         ▼                                 │
┌─────────────────┐                        │
│  Query MySQL    │                        │
│    Database     │                        │
└────────┬────────┘                        │
         │                                 │
         ▼                                 │
┌─────────────────┐                        │
│  Store in Redis │                        │
│  (with 10m TTL) │                        │
└────────┬────────┘                        │
         │                                 │
         └─────────────────────────────────┘
                         │
                         ▼
                 ┌──────────────┐
                 │ Return Data  │
                 └──────────────┘
```

#### Cache Invalidation

Cache secara otomatis di-invalidate (dihapus) pada event berikut:

- **Update Product**: Menghapus cache `product:detail:{id}` dan semua cache list (`products:list:*`)
- **Delete Product**: Menghapus cache `product:detail:{id}` dan semua cache list
- **Create Product**: Menghapus semua cache list untuk memastikan produk baru muncul

#### Performance Optimization

- **Singleflight Pattern**: Mencegah **cache stampede** dengan memastikan hanya satu goroutine yang melakukan query database untuk key yang sama pada saat bersamaan.
- **Concurrent-Safe**: `CacheWrapper` aman digunakan oleh multiple goroutines.

---

## 📨 Event-Driven Architecture

### Apache Kafka Integration

Aplikasi ini menggunakan **Apache Kafka** sebagai message broker untuk menangani proses asinkron dan meningkatkan skalabilitas sistem.

#### Event Flow

```
┌──────────────────┐
│ Create Transaction│
│   (HTTP POST)     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Save to MySQL DB │
│  (Transactional) │
└────────┬─────────┘
         │ Success
         ▼
┌──────────────────┐
│ Publish Event to │
│      Kafka       │──────┐
└────────┬─────────┘      │
         │                │
         ▼                │
┌──────────────────┐      │
│ Return Response  │      │
│   to Client      │      │
└──────────────────┘      │
                          │
         ┌────────────────┘
         │ Async Processing
         ▼
┌──────────────────┐
│ Kafka Consumer   │
│ (Background Job) │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Send Email /    │
│  Notification    │
└──────────────────┘
```

#### Kafka Topics & Events

| Topic           | Event Type          | Producer             | Consumer (Future)      | Description                          |
| --------------- | ------------------- | -------------------- | ---------------------- | ------------------------------------ |
| `order_created` | `OrderCreatedEvent` | `TransactionService` | `NotificationConsumer` | Dipublish saat transaksi baru dibuat |

#### Event Payload: `OrderCreatedEvent`

```json
{
  "order_id": 123,
  "user_id": 456,
  "user_email": "user@example.com",
  "total_price": 150000.0,
  "created_at": "2024-12-06T14:30:00Z"
}
```

#### Why Kafka?

- ⚡ **Decoupling**: Service tidak perlu menunggu proses lambat (email, notification) selesai
- 🚀 **Scalability**: Consumer bisa di-scale secara independen
- 🔄 **Reliability**: Message tersimpan di Kafka sampai berhasil di-consume
- 📊 **Event Sourcing**: Log semua event penting untuk audit dan analytics

---

## 🧪 Testing (Comprehensive Suite)

Aplikasi ini dilengkapi dengan pengujian level-industri (*enterprise-grade unit testing*) yang menyimulasikan berbagai kondisi, seperti interupsi koneksi, pembatalan *context*, serta *cache miss* tanpa membebani *production environment*.

### 🏗️ Tools & Mocks yang Digunakan
1. **[testify/mock](https://github.com/stretchr/testify):** Digunakan pada *Service Layer* dan *Handler Layer* untuk *mocking* repository interface dan *service interface* secara akurat tanpa menyentuh *database* atau Redis asli.
2. **[DATA-DOG/go-sqlmock](https://github.com/DATA-DOG/go-sqlmock):** Digunakan untuk melakukan *mock* terhadap level driver SQL dan melacak query eksekusi `t.DB.BeginTx(ctx, nil)` serta konektivitas `Ping()`.
3. **httptest:** Standar library `net/http/httptest` dimanfaatkan dalam lapisan *Delivery* (`middleware` & `handler`) untuk merekam skenario HTTP (401 Unauthorized, 200 OK, Canceled Context, dsb.).

### ▶️ Run the Test Suite

```bash
# Jalankan seluruh skenario Unit Tests
go test -v ./...

# Jalankan Test dengan menampilkan Real Code Coverage dari seluruh package
go test -v -coverpkg=./... ./...

# Jalankan skenario per sub-package / domain specific (Contoh: area service validation)
go test -v ./test/service/...
```

### ✨ Contoh Beberapa Skema Evaluasi:
- **Resiliency Testing**: Simulasi Redis mati paksa melalui *invalid dummy port* (`localhost:9999`) untuk men-segera-kan *I/O timeout* memicu *fallback* ke sistem komputasi MySQL.
- **Context Cancellation**: Simulasi `context.WithCancel()` secara spesifik pada HTTP request handlers.
- **Strict Data-Driven**: Logika penetapan *Transaction Total Price* dijalankan murni oleh *backend pricing query* tanpa bisa dipengaruhi memanipulasi parameter di level JWT/Frontend.
---

## 🐛 Troubleshooting

<details>
<summary><b>Common Issues & Solutions</b></summary>

### Issue: Container gagal start

**Solution**:

```bash
# Stop semua container
docker-compose down

# Remove volumes (⚠️ ini akan menghapus data!)
docker-compose down -v

# Rebuild dan start ulang
docker-compose up --build
```

### Issue: Port already in use

**Solution**:

```bash
# Check port usage (Windows)
netstat -ano | findstr :3001
netstat -ano | findstr :9092

# Kill process atau ubah port di .env dan docker-compose.yml
```

### Issue: Kafka broker not reachable

**Solution**:

```bash
# Check Kafka container logs
docker logs kafka

# Verify Kafka is listening
docker exec -it kafka kafka-topics.sh --bootstrap-server localhost:9092 --list

# Check Zookeeper health
docker exec -it zookeeper zkServer.sh status
```

### Issue: Redis connection refused

**Solution**:

```bash
# Check Redis container
docker logs redis-cache-service

# Test Redis connection
docker exec -it redis-cache-service redis-cli
> AUTH redissecret123
> PING
```

### Issue: Database migration failed

**Solution**:

```bash
# Check migration status
docker exec -it go-app-service /bin/sh
migrate -database "$DB_URL" -path db/migrations version

# Force specific version (⚠️ hati-hati!)
migrate -database "$DB_URL" -path db/migrations force <version>
```

</details>

---

## 🚦 Development

### Local Development (without Docker)

<details>
<summary><b>Setup for local development</b></summary>

#### Prerequisites

- Go 1.22+
- MySQL 8.0
- Redis 7.0
- Apache Kafka 3.0+

#### Steps

1. Install dependencies:

```bash
go mod download
```

2. Install Wire (untuk regenerate dependency injection):

```bash
go install github.com/google/wire/cmd/wire@latest
```

3. Setup local MySQL, Redis, & Kafka

4. Update `.env` dengan local configuration:

```env
MYSQL_HOST=localhost
REDIS_HOST=localhost
KAFKA_BROKER=localhost:9092
```

5. Run migrations:

```bash
migrate -database "mysql://user:pass@tcp(localhost:3306)/dbname" -path db/migrations up
```

6. (Optional) Regenerate Wire code jika ada perubahan dependency:

```bash
cd cmd/injector
wire
```

7. Run application:

```bash
go run cmd/api/main.go
```

</details>

---

## 🗺️ Roadmap

- [ ] Implement Kafka Consumer untuk notifikasi email
- [ ] Add Prometheus metrics untuk monitoring
- [ ] Implement rate limiting middleware
- [ ] Add comprehensive integration tests
- [ ] Setup CI/CD pipeline (GitHub Actions)
- [ ] Add Swagger auto-generation
- [ ] Implement gRPC endpoints untuk inter-service communication

---

## 🤝 Kontribusi

Kontribusi sangat diterima! Silakan buka issue atau pull request untuk improvement.

### 📝 How to Contribute

1. Fork repository ini
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

## 👨‍💻 Author

**Ahmad Sofi Sidik**

[![LinkedIn](https://img.shields.io/badge/LinkedIn-Connect-0077B5?style=for-the-badge&logo=linkedin)](https://www.linkedin.com/in/ahmad-sofi-sidik/)
[![GitHub](https://img.shields.io/badge/GitHub-Follow-181717?style=for-the-badge&logo=github)](https://github.com/assidik12)

---

## 🌟 Show Your Support

Jika proyek ini membantu Anda, berikan ⭐️ di [GitHub](https://github.com/assidik12/go-restfull-api)!

---

<div align="center">

**[Back to Top ⬆️](#-go-e-commerce-rest-api)**

Made with ❤️ using Go • Powered by Clean Architecture

</div>
