# GO_comerce — Improvement & AI Integration Roadmap

> **Project:** `github.com/assidik12/GO_comerce`  
> **Stack:** Go 1.24 · MySQL · Redis · Kafka · Docker  
> **Goal:** Refactor ke production-grade codebase + integrasi AI Agent layer  
> **Total Estimasi:** 11–17 hari kerja konsisten

---

## Executive Summary

| Kategori | Status Sekarang | Target |
|---|---|---|
| Architecture | Clean Architecture ada tapi ada violation | Fully compliant Clean Architecture |
| Security | Total price dari client (exploitable) | Server-side calculation + stock check |
| Error Handling | Semua error → HTTP 500/404 membabi buta | Typed errors → HTTP status akurat |
| Kafka | Ter-inject tapi tidak pernah dipanggil | Event published setiap transaksi |
| Testing | 1 connection test | ≥60% coverage di service layer |
| Observability | `log.Println` plain text | Structured `slog` + Graceful shutdown |
| AI Features | Tidak ada | Smart recommendation + semantic search |

---

## Phase Overview

```
P1 Foundation Fix      ████████░░░░░░░░░░░░  1–2 hari   CRITICAL
P2 Security Fix        ████████████░░░░░░░░  2–3 hari   HIGH
P3 Reliability         ████████████████░░░░  2–3 hari   MEDIUM
P4 Testing             ██████████████████░░  3–4 hari   MEDIUM
P5 AI Agent Layer      ████████████████████  3–5 hari   UPGRADE
```

---

## Phase 1 — Foundation Fix

**Durasi:** 1–2 hari  
**Priority:** 🔴 CRITICAL — harus selesai sebelum phase lain

### 1.1 Pindah Repository Interface ke Domain Layer

**Why:** Interface `ProductRepository`, `UserRepository`, `TransactionRepository` sekarang ada di dalam package `mysql`. Ini berarti service layer depend langsung ke detail implementasi DB — melanggar Clean Architecture.

**Before:**
```go
// internal/repository/mysql/product.repository.go
package mysql

// ❌ Interface ada di layer implementasi
type ProductRepository interface {
    GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, error)
    FindById(ctx context.Context, id int) (domain.Product, error)
    Save(ctx context.Context, product domain.Product) (domain.Product, error)
}

// Service bergantung ke package mysql secara langsung
import "github.com/assidik12/go-restfull-api/internal/repository/mysql"

type productService struct {
    ProductRepository mysql.ProductRepository // ← depend ke mysql package
}
```

**After:**
```go
// internal/domain/port.go  ← FILE BARU
package domain

// ✅ Interface di domain layer, bebas dari detail implementasi
type ProductRepository interface {
    GetAll(ctx context.Context, page, pageSize int) ([]Product, error)
    FindById(ctx context.Context, id int) (Product, error)
    Save(ctx context.Context, product Product) (Product, error)
    Update(ctx context.Context, product Product) (Product, error)
    Delete(ctx context.Context, id int) error
}

type UserRepository interface {
    Save(ctx context.Context, user User) (User, error)
    FindByEmail(ctx context.Context, email string) (User, error)
    FindById(ctx context.Context, id int) (User, error)
}

// Service sekarang depend ke domain, bukan mysql
import "github.com/assidik12/go-restfull-api/internal/domain"

type productService struct {
    repo domain.ProductRepository // ← clean dependency
}
```

**Output:** Domain layer benar-benar independen dari DB. Bisa swap MySQL → PostgreSQL tanpa ubah satu baris pun di service layer.

---

### 1.2 Custom Error Types

**Why:** Handler tidak bisa bedain jenis error dari service. DB down, data tidak ketemu, input salah — semua dijadiin satu response type yang sama.

**Before:**
```go
// handler/product.go
product, err := ph.service.GetProductById(r.Context(), idInt)
if err != nil {
    response.NotFound(w, err.Error()) // ❌ DB down pun jadi 404
    return
}
```

**After:**
```go
// internal/domain/errors.go  ← FILE BARU
package domain

import "errors"

var (
    ErrNotFound     = errors.New("resource not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrUnauthorized = errors.New("unauthorized")
    ErrConflict     = errors.New("resource already exists")
)

// service/product.service.go
func (p *productService) GetProductById(ctx context.Context, id int) (domain.Product, error) {
    product, err := p.repo.FindById(ctx, id)
    if errors.Is(err, sql.ErrNoRows) {
        return domain.Product{}, fmt.Errorf("%w: product id %d", domain.ErrNotFound, id)
    }
    return product, err
}

// handler/product.go — sekarang bisa bedain error
func (ph *ProductHandler) GetProductById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    product, err := ph.service.GetProductById(r.Context(), idInt)
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrNotFound):
            response.NotFound(w, err.Error())            // 404
        case errors.Is(err, domain.ErrInvalidInput):
            response.BadRequest(w, err.Error())          // 400
        default:
            response.InternalServerError(w, "internal error") // 500
        }
        return
    }
    response.OK(w, product)
}
```

**Output:** HTTP status code akurat. API response lebih informatif. Client bisa handle error dengan benar.

---

### 1.3 Fix Typos & Naming Convention

**Why:** Typo di nama struct/interface tersebar di codebase — unprofessional dan bikin `golint`/`staticcheck` warning.

**Before:**
```go
// ❌ Semua typo ini tersebar di codebase
type TrancationService interface { ... }  // harusnya TransactionService
Quantyty   int                           // harusnya Quantity
Transaction_detail_id string            // harusnya TransactionDetailID (Go convention)
User_id   int                           // harusnya UserID
Total_Price int                         // harusnya TotalPrice
```

**After:**
```go
// ✅ Naming sesuai Go convention — no snake_case di struct fields
type TransactionService interface { ... }

type Transaction struct {
    ID                  int
    TransactionDetailID string
    UserID              int
    TotalPrice          int
    Products            []TransactionDetail
}

type TransactionDetail struct {
    Quantity  int
    ProductID int
}
```

**Output:** Codebase bersih dari lint warning. Konsisten dengan Go standard naming convention.

---

## Phase 2 — Security & Correctness

**Durasi:** 2–3 hari  
**Priority:** 🟠 HIGH

### 2.1 Server-Side Price Calculation

**Why:** `TotalPrice` sekarang dikirim dari client. Ini bug kritis — user bisa kirim harga Rp1 untuk produk Rp500.000. Tidak ada yang memvalidasi.

**Before:**
```go
// dto/transaction.go
type TransactionRequest struct {
    Products   []ProductItem `json:"products"`
    TotalPrice int           `json:"total_price"` // ❌ BAHAYA: client tentukan harga
}

// service/transaction.service.go
transactionToSave := domain.Transaction{
    TotalPrice: transaction.TotalPrice, // ← trust client blindly
}
```

**After:**
```go
// dto/transaction.go
type TransactionRequest struct {
    Products []ProductItem `json:"products"` // ✅ hanya kirim produk & qty
    // TotalPrice DIHAPUS dari request
}

// service/transaction.service.go
func (t *transactionService) Save(ctx context.Context, req dto.TransactionRequest, userID int) (domain.Transaction, error) {
    // Hitung harga server-side
    var totalPrice int
    for _, item := range req.Products {
        product, err := t.productRepo.FindById(ctx, item.ProductID)
        if err != nil {
            return domain.Transaction{}, fmt.Errorf("%w: product id %d", domain.ErrNotFound, item.ProductID)
        }
        if product.Stock < item.Quantity {
            return domain.Transaction{}, fmt.Errorf("%w: insufficient stock for product %d", domain.ErrInvalidInput, item.ProductID)
        }
        totalPrice += product.Price * item.Quantity
    }

    transactionToSave := domain.Transaction{
        UserID:     userID,
        TotalPrice: totalPrice, // ← dihitung server, bukan dari client
    }
    // ...
}
```

**Output:** Harga tidak bisa dimanipulasi. Stock otomatis dicek sebelum transaksi disimpan.

---

### 2.2 Hapus Config Coupling di Service Layer

**Why:** `userService.Login()` memanggil `config.GetConfig()` langsung di dalam service — service layer tidak boleh akses global config. Ini bikin unit test susah dan coupling tinggi.

**Before:**
```go
// service/user.service.go
func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
    cfg := config.GetConfig() // ❌ service depend ke config package global
    // ...
    token, err := jwt.NewJWTService(cfg.JWTSecret).GenerateJWT(user)
}

func NewUserService(repo mysql.UserRepository, DB *sql.DB, validate *validator.Validate) UserService {
    return &userService{repo: repo, DB: DB, validate: validate}
}
```

**After:**
```go
// service/user.service.go
type userService struct {
    repo      domain.UserRepository
    DB        *sql.DB
    validate  *validator.Validate
    jwtSecret string // ✅ di-inject lewat constructor
}

func NewUserService(
    repo domain.UserRepository,
    DB *sql.DB,
    validate *validator.Validate,
    jwtSecret string, // ← inject, bukan ambil dari global config
) UserService {
    return &userService{
        repo:      repo,
        DB:        DB,
        validate:  validate,
        jwtSecret: jwtSecret,
    }
}

func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
    // ...
    token, err := jwt.NewJWTService(s.jwtSecret).GenerateJWT(user) // ✅ pakai field
}
```

**Output:** Service layer murni — tidak ada global state. Mudah di-unit test dengan inject secret mock.

---

### 2.3 Fix SQL Issues & Resource Leaks

**Why:** `SELECT *` rawan break kalau kolom DB berubah urutan. `rows.Close()` manual bisa leak kalau ada early return atau panic.

**Before:**
```go
// repository/mysql/product.repository.go

// ❌ SELECT * — urutan kolom bisa berubah kalau schema berubah
q := "SELECT * FROM products WHERE id = ?"

// ❌ rows.Close() manual — bisa leak kalau ada early return
rows, err := p.db.QueryContext(ctx, query, pageSize, offset)
// ... loop ...
rows.Close()
return products, nil
```

**After:**
```go
// repository/mysql/product.repository.go

// ✅ Explicit columns — aman dari perubahan schema
q := `SELECT id, name, price, stock, description, img, category_id
      FROM products WHERE id = ?`

err := p.db.QueryRowContext(ctx, q, id).Scan( // ✅ QueryRowContext (context-aware)
    &product.ID, &product.Name, &product.Price,
    &product.Stock, &product.Description, &product.Img, &product.CategoryID,
)
if errors.Is(err, sql.ErrNoRows) {
    return domain.Product{}, domain.ErrNotFound
}

// ✅ defer rows.Close() — aman dari panic / early return
rows, err := p.db.QueryContext(ctx, query, pageSize, offset)
if err != nil {
    return nil, err
}
defer rows.Close()

for rows.Next() {
    // ...
}
return products, rows.Err() // ✅ cek error dari iterasi
```

**Output:** Tidak ada resource leak. Query aman dari perubahan schema. Context cancellation berjalan dengan benar.

---

## Phase 3 — Reliability & Observability

**Durasi:** 2–3 hari  
**Priority:** 🟡 MEDIUM

### 3.1 Graceful Shutdown

**Why:** Server sekarang langsung mati saat menerima SIGTERM. In-flight request terpotong. Koneksi DB/Redis tidak ditutup bersih.

**Before:**
```go
// cmd/api/main.go
func main() {
    server, cleanup, err := injector.InitializedServer(*cfg)
    defer cleanup()

    server.Addr = fmt.Sprintf(":%s", cfg.AppPort)

    // ❌ SIGTERM langsung kill proses, request yang sedang diproses terpotong
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

**After:**
```go
// cmd/api/main.go
func main() {
    cfg := config.GetConfig()
    server, cleanup, err := injector.InitializedServer(*cfg)
    if err != nil {
        log.Fatalf("Failed to initialize server: %v", err)
    }
    defer cleanup()

    server.Addr = fmt.Sprintf(":%s", cfg.AppPort)

    // Jalankan server di goroutine terpisah
    go func() {
        log.Printf("Server starting on port %s...", cfg.AppPort)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // ✅ Tunggu SIGINT atau SIGTERM
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down gracefully...")

    // Beri waktu 30 detik untuk in-flight request selesai
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Forced shutdown: %v", err)
    }
    log.Println("Server exited cleanly.")
}
```

**Output:** Zero downtime deployment. In-flight request selesai sebelum server mati. DB/Redis connection ditutup bersih.

---

### 3.2 Structured Logging dengan `slog`

**Why:** Sekarang pakai `log.Println` biasa. Tidak ada level, tidak ada structured fields, tidak bisa di-query di cloud logging platform.

**Before:**
```go
// Di berbagai tempat — plain text, tidak terstruktur
log.Println("connection to kafka success...")
log.Printf("[Kafka] Failed to publish to topic %s: %v", topic, err)
log.Printf("Server GO_comerce is starting on port %s...", cfg.AppPort)
```

**After:**
```go
// internal/pkg/logger/logger.go  ← FILE BARU
package logger

import (
    "log/slog"
    "os"
)

func New(env string) *slog.Logger {
    var handler slog.Handler
    if env == "production" {
        // JSON untuk production — bisa di-query di Datadog/Loki/CloudWatch
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    } else {
        // Text untuk development — human-readable
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }
    return slog.New(handler)
}

// Pemakaian di event/kafka.go
func (k *KafkaProducer) Publish(ctx context.Context, topic string, message interface{}) error {
    // ...
    if err != nil {
        k.logger.Error("kafka publish failed",
            "topic", topic,
            "error", err,
        )
        return err
    }
    k.logger.Info("kafka message published", "topic", topic)
    return nil
}
```

**Output:** Log terstruktur dengan level. Siap untuk centralized logging (Loki, Datadog, CloudWatch).

---

### 3.3 Complete Kafka Event Publishing

**Why:** Kafka producer di-inject ke `TransactionService` tapi tidak pernah dipanggil. Fitur event-driven-nya terpasang tapi tidak berfungsi.

**Before:**
```go
// service/transaction.service.go
func (t *transactionService) Save(ctx context.Context, ...) (domain.Transaction, error) {
    // ... simpan ke DB ...
    if err := tx.Commit(); err != nil {
        return domain.Transaction{}, err
    }

    // ❌ Producer di-inject tapi tidak pernah dipakai
    return savedTransaction, nil
}
```

**After:**
```go
// internal/event/order.event.go — tambah event struct
package event

const TopicOrderCreated = "order.created"

type OrderCreatedEvent struct {
    TransactionID string    `json:"transaction_id"`
    UserID        int       `json:"user_id"`
    TotalPrice    int       `json:"total_price"`
    Products      []Product `json:"products"`
    CreatedAt     time.Time `json:"created_at"`
}

// service/transaction.service.go
func (t *transactionService) Save(ctx context.Context, ...) (domain.Transaction, error) {
    // ... simpan ke DB, hitung harga server-side ...

    if err := tx.Commit(); err != nil {
        return domain.Transaction{}, err
    }

    // ✅ Publish event setelah commit berhasil — async agar tidak block response
    orderEvent := event.OrderCreatedEvent{
        TransactionID: savedTransaction.TransactionDetailID,
        UserID:        savedTransaction.UserID,
        TotalPrice:    savedTransaction.TotalPrice,
        CreatedAt:     time.Now(),
    }

    go func() {
        pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := t.producer.Publish(pubCtx, event.TopicOrderCreated, orderEvent); err != nil {
            t.logger.Error("failed to publish order event", "error", err)
        }
    }()

    return savedTransaction, nil
}
```

**Output:** Event-driven architecture aktif. Order notification, inventory update, email konfirmasi bisa di-trigger dari consumer Kafka.

---

## Phase 4 — Testing

**Durasi:** 3–4 hari  
**Priority:** 🟡 MEDIUM

### 4.1 Unit Test Service Layer dengan Mock

**Why:** Sekarang hanya ada 1 connection test. Tidak ada test untuk business logic. Bug bisa masuk tanpa ada yang tahu.

**Before:**
```go
// test/connection.test.go — satu-satunya test yang ada
func TestMySQLConnection(t *testing.T) {
    db, err := sql.Open("mysql", dsn)
    assert.NoError(t, err)
    assert.NoError(t, db.Ping())
}
// ❌ Tidak ada test untuk service logic
// ❌ Tidak ada mock untuk repository
// ❌ Business rule tidak terverifikasi sama sekali
```

**After:**
```go
// internal/service/product_service_test.go  ← FILE BARU
package service_test

import (
    "context"
    "errors"
    "testing"

    "github.com/assidik12/go-restfull-api/internal/domain"
    "github.com/assidik12/go-restfull-api/internal/service"
    "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock ProductRepository
type MockProductRepo struct{ mock.Mock }

func (m *MockProductRepo) FindById(ctx context.Context, id int) (domain.Product, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(domain.Product), args.Error(1)
}

// ... implement semua method interface ...

func TestGetProductById_Success(t *testing.T) {
    mockRepo := new(MockProductRepo)
    mockCache := new(MockCache)

    expected := domain.Product{ID: 1, Name: "Laptop", Price: 10_000_000}
    mockRepo.On("FindById", mock.Anything, 1).Return(expected, nil)
    mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("miss"))
    mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

    svc := service.NewProductService(mockRepo, nil, mockCache, validator.New())
    result, err := svc.GetProductById(context.Background(), 1)

    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockRepo.AssertExpectations(t)
}

func TestGetProductById_NotFound(t *testing.T) {
    mockRepo := new(MockProductRepo)
    mockCache := new(MockCache)

    mockRepo.On("FindById", mock.Anything, 99).Return(domain.Product{}, domain.ErrNotFound)
    mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("miss"))

    svc := service.NewProductService(mockRepo, nil, mockCache, validator.New())
    _, err := svc.GetProductById(context.Background(), 99)

    assert.ErrorIs(t, err, domain.ErrNotFound) // ✅ verifikasi error type spesifik
}

func TestSaveTransaction_ServerSidePriceCalculation(t *testing.T) {
    // Verifikasi bahwa harga dihitung server-side, tidak dari request
    mockProductRepo := new(MockProductRepo)
    product := domain.Product{ID: 1, Price: 500_000, Stock: 10}
    mockProductRepo.On("FindById", mock.Anything, 1).Return(product, nil)

    req := dto.TransactionRequest{
        Products: []dto.ProductItem{{ProductID: 1, Quantity: 2}},
        // TotalPrice tidak ada di request — dihitung server
    }

    result, err := svc.Save(context.Background(), req, userID)
    assert.NoError(t, err)
    assert.Equal(t, 1_000_000, result.TotalPrice) // 500.000 × 2
}
```

**Output:** Business logic tercover test. Regression terlindungi. CI/CD bisa auto-reject PR yang break business rule.

---

## Phase 5 — AI Agent Layer

**Durasi:** 3–5 hari  
**Priority:** 🟣 UPGRADE — differentiator utama

### 5.1 AI-Powered Product Recommendation & Smart Search

**Why:** Ini differentiator utama di 2025. Natural language search jauh lebih baik dari pagination biasa. Bisa jadi showcase di portfolio untuk masuk S1.

**Before:**
```go
// Sekarang: tidak ada search sama sekali
// GET /api/v1/products — hanya pagination biasa
// User harus manual scroll semua produk
// Tidak ada filter, tidak ada rekomendasi
```

**After:**
```go
// internal/service/ai.service.go  ← FILE BARU
package service

type AIService interface {
    Recommend(ctx context.Context, query string, products []domain.Product) (dto.AIRecommendResponse, error)
}

type anthropicAIService struct {
    apiKey     string
    httpClient *http.Client
    logger     *slog.Logger
}

func (s *anthropicAIService) Recommend(ctx context.Context, query string, products []domain.Product) (dto.AIRecommendResponse, error) {
    systemPrompt := `Kamu adalah asisten belanja yang helpful.
Berdasarkan katalog produk yang diberikan, rekomendasikan produk yang paling sesuai dengan query user.
Respond ONLY dalam format JSON:
{"recommendations": [{"product_id": int, "reason": "alasan singkat dalam bahasa Indonesia"}]}`

    userMessage := fmt.Sprintf(
        "Query user: %s\n\nKatalog Produk:\n%s",
        query, formatProductsJSON(products),
    )

    // Panggil Anthropic API
    resp, err := s.callAnthropicAPI(ctx, systemPrompt, userMessage)
    if err != nil {
        return dto.AIRecommendResponse{}, err
    }

    return parseRecommendations(resp, products)
}

// internal/delivery/http/handler/ai.handler.go  ← FILE BARU
// POST /api/v1/ai/recommend
// Body: { "query": "laptop untuk coding bawah 10 juta" }
func (h *AIHandler) SmartRecommend(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    var req dto.AIRecommendRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    // Ambil semua produk sebagai context
    products, err := h.productService.GetAllProducts(r.Context(), 1, 100)
    if err != nil {
        response.InternalServerError(w, err.Error())
        return
    }

    result, err := h.aiService.Recommend(r.Context(), req.Query, products)
    if err != nil {
        response.InternalServerError(w, err.Error())
        return
    }

    response.OK(w, result)
}

// Routes baru yang ditambahkan:
// POST /api/v1/ai/recommend      ← "cari laptop gaming bawah 15 juta"
// POST /api/v1/ai/search         ← semantic search (bukan exact match keyword)
// GET  /api/v1/ai/product/:id/summary ← AI summary dari deskripsi produk
```

**Output:** Natural language search yang benar-benar bekerja. Bisa jadi fitur utama di portfolio. Mudah di-extend ke chatbot customer service.

---

### 5.2 Rate Limiting & Middleware Chain

**Why:** API tidak ada rate limiting. Endpoint AI yang ada cost per-request bisa di-spam sampai biaya meledak.

**Before:**
```go
// routes.go — tidak ada proteksi apapun
router.POST("/api/v1/ai/recommend", aiHandler.SmartRecommend)
// ❌ Siapapun bisa spam → biaya Anthropic API tidak terkontrol
// ❌ Tidak ada timeout per-request
// ❌ Tidak ada request ID untuk tracing
```

**After:**
```go
// internal/delivery/http/middleware/chain.go  ← FILE BARU
package middleware

// Compose beberapa middleware jadi satu
func Chain(next httprouter.Handle, middlewares ...func(httprouter.Handle) httprouter.Handle) httprouter.Handle {
    for i := len(middlewares) - 1; i >= 0; i-- {
        next = middlewares[i](next)
    }
    return next
}

// Rate limiter per IP menggunakan token bucket (stdlib)
func RateLimit(rps int, burst int) func(httprouter.Handle) httprouter.Handle {
    limiters := &sync.Map{}

    return func(next httprouter.Handle) httprouter.Handle {
        return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
            ip := realIP(r)
            limiter, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rate.Limit(rps), burst))

            if !limiter.(*rate.Limiter).Allow() {
                http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
                return
            }
            next(w, r, ps)
        }
    }
}

// routes.go — middleware chain per endpoint
router.POST("/api/v1/ai/recommend",
    Chain(
        aiHandler.SmartRecommend,
        middleware.RequestID(),                   // inject X-Request-ID ke setiap request
        middleware.RequestLogger(logger),         // log method, path, latency, status
        middleware.RateLimit(5, 10),             // max 5 req/s, burst 10 per IP
        middleware.Timeout(30*time.Second),       // timeout 30s (AI bisa lambat)
        middleware.Auth("user", cfg.JWTSecret),  // harus login
    ),
)
```

**Output:** Biaya AI API terkontrol. Setiap request punya ID untuk tracing. API terlindungi dari abuse.

---

## File Structure — Before vs After

```
Before:
go-restfull-api/
├── internal/
│   ├── domain/
│   │   ├── product.go          ← hanya struct
│   │   ├── user.go             ← hanya struct
│   │   └── transaction.go      ← hanya struct, ada typo
│   ├── repository/
│   │   └── mysql/
│   │       ├── product.repository.go   ← interface + implementasi campur
│   │       ├── user.repository.go      ← interface + implementasi campur
│   │       └── transaction.repository.go
│   ├── service/
│   │   ├── product.service.go  ← depend ke mysql package
│   │   ├── user.service.go     ← akses config global
│   │   └── trancation.service.go  ← typo, producer tidak dipakai
│   └── ...
└── test/
    └── connection.test.go      ← satu-satunya test

After:
go-restfull-api/
├── internal/
│   ├── domain/
│   │   ├── product.go
│   │   ├── user.go
│   │   ├── transaction.go      ← naming fixed
│   │   ├── port.go             ← ✨ BARU: semua repository interface
│   │   └── errors.go           ← ✨ BARU: typed errors
│   ├── pkg/
│   │   └── logger/
│   │       └── logger.go       ← ✨ BARU: structured slog
│   ├── repository/
│   │   └── mysql/
│   │       ├── product.repository.go   ← hanya implementasi
│   │       ├── user.repository.go      ← hanya implementasi
│   │       └── transaction.repository.go
│   ├── service/
│   │   ├── product.service.go  ← depend ke domain.ProductRepository
│   │   ├── user.service.go     ← jwtSecret di-inject
│   │   ├── transaction.service.go  ← naming fixed, Kafka dipakai
│   │   └── ai.service.go       ← ✨ BARU: Anthropic AI integration
│   ├── delivery/http/
│   │   ├── handler/
│   │   │   └── ai.handler.go   ← ✨ BARU
│   │   └── middleware/
│   │       └── chain.go        ← ✨ BARU: rate limit, timeout, request ID
│   └── event/
│       └── order.event.go      ← ✨ event struct + TopicOrderCreated
├── test/
│   ├── connection.test.go
│   └── service/
│       ├── product_service_test.go     ← ✨ BARU
│       ├── user_service_test.go        ← ✨ BARU
│       └── transaction_service_test.go ← ✨ BARU
└── cmd/api/main.go             ← graceful shutdown
```

---

## Timeline

```
Minggu 1
├── Hari 1–2   P1: Foundation Fix (interface, error types, naming)
├── Hari 3–4   P2: Security Fix (price calculation, config, SQL)
└── Hari 5     P3: Graceful shutdown + slog

Minggu 2
├── Hari 6–7   P3: Kafka event publishing
├── Hari 8–9   P4: Unit test product & user service
└── Hari 10    P4: Unit test transaction service

Minggu 3
├── Hari 11–12 P5: AI service + Anthropic API integration
├── Hari 13–14 P5: AI handler + routes + rate limiting
└── Hari 15–17 P5: Testing, polish, dokumentasi
```

---

## Quick Reference — Issues vs Fix

| # | Issue | Lokasi | Fix | Phase |
|---|---|---|---|---|
| 1 | Repository interface di mysql package | `repository/mysql/*.go` | Pindah ke `domain/port.go` | P1 |
| 2 | Tidak ada typed errors | semua handler | Buat `domain/errors.go` | P1 |
| 3 | Typo `TrancationService`, `Quantyty` | service, domain | Rename semua | P1 |
| 4 | TotalPrice dari client | `service/transaction.service.go` | Hitung server-side + stock check | P2 |
| 5 | `config.GetConfig()` di service | `service/user.service.go` | Inject `jwtSecret` via constructor | P2 |
| 6 | `SELECT *` di query | `repository/mysql/product.repository.go` | Explicit column list | P2 |
| 7 | `rows.Close()` manual | `repository/mysql/product.repository.go` | `defer rows.Close()` | P2 |
| 8 | Tidak ada graceful shutdown | `cmd/api/main.go` | `signal.Notify` + `server.Shutdown` | P3 |
| 9 | `log.Println` plain text | semua file | Structured `slog` | P3 |
| 10 | Kafka producer tidak dipanggil | `service/transaction.service.go` | Async publish setelah commit | P3 |
| 11 | Tidak ada unit test | `test/` | Mock-based service tests | P4 |
| 12 | Tidak ada AI feature | — | Anthropic API + smart search | P5 |
| 13 | Tidak ada rate limiting | `routes.go` | Token bucket per IP + middleware chain | P5 |

---

*Generated for GO_comerce project — April 2026*
