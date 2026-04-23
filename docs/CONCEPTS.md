# Key Concepts & Design Decisions

## 1. Why Clean Architecture?

**Problem:** Tightly coupled code is hard to test, maintain, and scale.

**Solution:** Organize code in layers where each layer only knows about 
layers below it. Dependencies flow inward (toward domain).

```
Presentation Layer (handlers) ↘
Business Logic Layer (service) → Domain (entities, rules)
Data Access Layer (repository) ↗
```

**Benefit:** Change database from MySQL to PostgreSQL? Only update 
repository. Service and handler code unchanged.

---

## 2. Event-Driven Architecture

**Problem:** E-commerce needs multiple side effects:
- Send confirmation email
- Update inventory
- Create shipping label
- Calculate loyalty points

Doing this synchronously makes transactions slow and fragile.

**Solution:** After transaction commits to DB, publish event to Kafka. 
Consumers handle side effects async.

```
Benefits:
✅ Transaction completes quickly (no email/shipping delay)
✅ Reliable (Kafka retries if service down)
✅ Scalable (add new consumer without changing core)
✅ Decoupled (email service doesn't need to know about shipping)
```

---

## 3. Cache Stampede Prevention

**Problem:** 1000 concurrent requests for same product = 1000 DB hits

**Solution:** Use singleflight to ensure only 1 goroutine queries DB, 
other 999 wait for result.

```
Without singleflight:
1000 requests → 1000 DB queries → Database overload

With singleflight:
1000 requests → 1 DB query → 1000 responses from same data
```

---

## 4. Transaction Boundaries

**Wrong:**
```go
// Each operation in separate transaction
user := getUserById()        // Transaction 1
product := getProduct()      // Transaction 2
decrementStock()             // Transaction 3 ← If crashes here, previous operations already committed
```

**Right (Catalyst):**
```go
tx := db.Begin()
  user := getUserById(tx)
  product := getProduct(tx)
  decrementStock(tx)
tx.Commit() // All succeed or all rollback
```

---

## 5. Error Semantics

**Problem:** Service returns `error` but handler doesn't know:
- Is it "not found" (404)?
- Is it "invalid input" (400)?
- Is it "internal error" (500)?

**Solution:** Use sentinel errors that handler can check:

```go
// Service returns
if err == sql.ErrNoRows {
    return product{}, fmt.Errorf("%w: ...", domain.ErrNotFound)
}

// Handler checks
case errors.Is(err, domain.ErrNotFound):
    http.NotFound(w, ...)  // 404
```

Now handler can return correct HTTP status.

---

## 6. Dependency Injection

**Problem:** Service needs jwtSecret but where does it get it?

**Option A (Bad):** Service imports config package
```go
// ❌ Service depends on external package
func (s *userService) Login() {
    cfg := config.GetConfig()  // Global access
    secret := cfg.JWTSecret
}
```

**Option B (Good - Catalyst):** Inject via constructor
```go
// ✅ Service depends on parameter
func NewUserService(jwtSecret string, ...) UserService {
    return &userService{jwtSecret: jwtSecret}
}
```

Benefits:
- Service is testable (inject mock secret)
- Service is reusable (can be embedded in different contexts)
- Dependencies are explicit (visible in constructor)

---

## 7. Structured Logging

**Problem:** Can't query plain text logs in production

```
2026-04-23 10:30 Starting server  // What port?
2026-04-23 10:31 Error            // What error?
```

**Solution (Catalyst):** JSON structured logs

```json
{"time":"2026-04-23T10:30:15Z","level":"INFO","msg":"Server starting","port":"8080"}
{"time":"2026-04-23T10:31:22Z","level":"ERROR","msg":"Transaction failed","transaction_id":"tx-123","error":"Insufficient stock"}
```

Now you can query: `SELECT * WHERE error contains "Insufficient stock"`

---

## 8. Graceful Shutdown

**Problem:** Server gets SIGTERM, immediately kills in-flight requests.

**Solution (Catalyst):**
```
Receive SIGTERM
  ↓
Stop accepting new connections
  ↓
Wait up to 30 seconds for in-flight requests to finish
  ↓
Force shutdown if timeout exceeded
  ↓
Close DB/Redis/Kafka connections cleanly
```

**Result:** Zero data loss, no corrupted transactions.
