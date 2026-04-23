# Technical Architecture — Catalyst

## Design Philosophy

Catalyst follows **Clean Architecture** + **Event-Driven Architecture** pattern, 
designed to handle real-world e-commerce constraints:

1. **Data Integrity** — Atomic transactions, no double-charging
2. **Performance** — Multi-layer caching with stampede prevention
3. **Reliability** — Graceful degradation, async fallback, message durability
4. **Maintainability** — Clear separation of concerns, testable layers

## Layered Architecture

```
┌─────────────────┐
│  HTTP Handler   │ ← Presentation (DTOs, error mapping)
├─────────────────┤
│    Service      │ ← Business Logic (calculations, validation)
├─────────────────┤
│  Repository     │ ← Data Access (SQL, caching)
├─────────────────┤
│  Infrastructure │ ← External Services (DB, Redis, Kafka)
└─────────────────┘
```

Each layer:
- **Depends only on the layer below** (or domain)
- **Exposes only what's necessary** (via interfaces)
- **Is independently testable** (via mocks)

## Event-Driven Processing

```
User Request → Service validates → DB transaction → Commit
                                        ↓
                        (If successful) → Publish to Kafka
                                        ↓
                        Async Consumer → Sends email, updates inventory, etc.
```

**Key principle:** The transaction commits FIRST, event publishes AFTER. 
This guarantees event durability (if Kafka fails, we can retry).

## Caching Strategy

```
Request → Check Redis
            ↓
         Cache hit? → Return immediately (5-10ms)
            ↓
         Cache miss? → Single flight prevents thundering herd
            ↓
         Query database → Cache result → Return
```

Singleflight prevents 1000 concurrent requests causing 1000 DB hits.
Instead: 1 DB hit, 999 requests wait for result, all get same answer.

## Performance Characteristics

| Scenario | Latency |
|---|---|
| Cached product detail | 5-10ms |
| DB hit (cache miss) | 50-150ms |
| Full transaction | 200-500ms |
| With Kafka publish | <100ms extra (async) |

---

## Transaction Atomicity

### The Problem
```
Transaction started
  ├─ Charge customer Rp500.000
  ├─ Decrement inventory by 1
  └─ Crash here! ← Customer charged but inventory never decremented
```

### The Solution (Catalyst)
```
Database transaction started
  ├─ Lock rows
  ├─ Charge customer Rp500.000
  ├─ Decrement inventory by 1
  ├─ All success? → Commit (atomic)
  └─ Any error? → Rollback (everything reverted)
  
Result: No half-state. Either all succeed or nothing happens.
```

---

## Server-Side Validation

### The Problem
```
Client sends: POST /transactions
{
  "products": [{"product_id": 1, "qty": 2}],
  "total_price": 100  // ← Attacker sets to Rp100 instead of Rp1.000.000
}
```

### The Solution (Catalyst)
```go
// Service calculates price from database
product := repo.FindById(1)     // Returns price: Rp500.000
totalPrice := product.Price * qty  // 500.000 × 2 = 1.000.000
// Client's total_price ignored entirely
```

Client cannot manipulate pricing. Server is the source of truth.
