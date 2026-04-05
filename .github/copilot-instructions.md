# Copilot Instructions — saweria-be (Golang Backend)

## Project Context

Backend untuk platform donasi streamer (Saweria Clone). Menerima donasi dari donor publik, memproses pembayaran via Midtrans, dan mengirim notifikasi real-time ke overlay OBS streamer via WebSocket.

## Tech Stack

- **Language:** Go (Golang)
- **HTTP Framework:** `gin-gonic/gin`
- **Database:** PostgreSQL
- **ORM / Query:** `jmoiron/sqlx` (raw SQL + named query, bukan ORM)
- **Auth:** JWT (`golang-jwt/jwt/v5`)
- **Payment:** Midtrans Go SDK (`veritrans/midtrans-go`)
- **WebSocket:** `gorilla/websocket`
- **Config:** `joho/godotenv`
- **Migration:** `golang-migrate/migrate`
- **Logging:** `log/slog` (stdlib)

## Struktur Folder

```
saweria-be/
├── cmd/
│   └── main.go               ← entry point, setup router & server
├── internal/
│   ├── auth/                 ← register, login, Google OAuth, refresh token
│   ├── user/                 ← profil streamer, update profil
│   ├── overlay/              ← alert settings, template, filter kata, suara, stream key
│   ├── donation/             ← submit donasi, histori donasi
│   ├── payment/              ← Midtrans Snap token creation, webhook handler
│   ├── wallet/               ← saldo streamer, cashout
│   └── websocket/            ← hub koneksi WS per stream_key
├── pkg/
│   ├── middleware/           ← JWT auth middleware, rate limiter
│   ├── database/             ← inisialisasi koneksi PostgreSQL
│   └── config/               ← load env variables
└── migrations/               ← SQL migration files
```

Setiap package `internal/*` mengikuti pola **3 layer**:

- `handler.go` — Gin handler: bind request, validasi input, panggil service, return JSON
- `service.go` — business logic murni; tidak boleh tahu soal HTTP atau SQL detail
- `repository.go` — query PostgreSQL via sqlx; tidak ada business logic

Setiap layer berkomunikasi melalui **interface** — service menerima interface repository, bukan concrete struct. Ini memudahkan mocking saat unit test.

## Konvensi Kode

### Handler (Gin)

```go
// Handler hanya bertanggung jawab pada HTTP layer
func (h *Handler) Create(c *gin.Context) {
    var req CreateRequest
    // ShouldBindJSON otomatis return 400 jika JSON malformed
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := h.service.Create(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"message": "ok", "data": result})
}
```

### Service Layer

```go
// Service interface — didefinisikan di package yang sama dengan handler
type Service interface {
    Create(ctx context.Context, req CreateRequest) (*Result, error)
}

// Repository interface — didefinisikan di package service/internal
type Repository interface {
    Insert(ctx context.Context, entity *Entity) (*Entity, error)
    FindByID(ctx context.Context, id string) (*Entity, error)
}

// Concrete service menerima interface repository
type service struct {
    repo Repository
}

func NewService(repo Repository) Service {
    return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Result, error) {
    // business logic di sini
    entity := &Entity{...}
    return s.repo.Insert(ctx, entity)
}
```

### Setup Router (Gin)

```go
// cmd/main.go
r := gin.Default()

// Middleware global
r.Use(middleware.CORS())
r.Use(gin.Recovery())

// Public routes
auth := r.Group("/auth")
{
    auth.POST("/register", authHandler.Register)
    auth.POST("/login", authHandler.Login)
}

// Protected routes
api := r.Group("/")
api.Use(middleware.JWTAuth(cfg.JWTSecret))
{
    api.GET("/users/me", userHandler.GetMe)
    api.PUT("/users/me", userHandler.UpdateMe)
    api.GET("/overlay/settings", overlayHandler.GetSettings)
    // ...
}

// WebSocket
r.GET("/ws/overlay", wsHandler.Handle)
```

### Response Format

Selalu kembalikan JSON konsisten menggunakan `c.JSON`:

```go
// Success
c.JSON(http.StatusOK, gin.H{"message": "ok", "data": result})

// Created
c.JSON(http.StatusCreated, gin.H{"message": "created", "data": result})

// Error
c.JSON(http.StatusBadRequest, gin.H{"error": "deskripsi error"})
c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
```

### Error Handling

- Jangan panik (`panic`) di handler atau service kecuali di `main.go` untuk fatal error saat startup
- Gunakan error wrapping: `fmt.Errorf("donation.Create: %w", err)`
- Log error di service layer, jangan di repository
- Kembalikan error yang bermakna ke handler, bukan error teknis database

### Database

- Semua query menggunakan parameter binding (`$1, $2`) — tidak pernah string concatenation untuk mencegah SQL injection
- Gunakan transaction untuk operasi yang melibatkan lebih dari satu tabel (contoh: buat donasi + kurangi saldo)
- Repository selalu menerima `context.Context` sebagai parameter pertama

### Auth & JWT

- JWT disimpan di `Authorization: Bearer <token>` header
- Payload JWT berisi: `user_id`, `email`, `username`, `exp`
- Middleware JWT (Gin middleware) mengekstrak `user_id` dan menyimpannya ke Gin context:

    ```go
    // Middleware
    func JWTAuth(secret string) gin.HandlerFunc {
        return func(c *gin.Context) {
            token := extractBearerToken(c)
            claims, err := validateToken(token, secret)
            if err != nil {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
                return
            }
            c.Set("user_id", claims.UserID)
            c.Set("username", claims.Username)
            c.Next()
        }
    }

    // Di handler, ambil dari context
    userID := c.GetString("user_id")
    ```

- Endpoint publik (halaman donasi, webhook Midtrans, WebSocket OBS) tidak memerlukan JWT

### WebSocket (gorilla/websocket + Gin)

```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, // sesuaikan di production
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

// Handler Gin untuk upgrade koneksi
func (h *WSHandler) Handle(c *gin.Context) {
    streamKey := c.Query("key")
    // validasi stream key ke DB
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    h.hub.Register(conn, streamKey)
    defer h.hub.Unregister(conn, streamKey)
    // read loop untuk keep-alive
    for {
        if _, _, err := conn.ReadMessage(); err != nil {
            break
        }
    }
}
```

- Hub (`websocket/hub.go`) mengelola `sync.Map` atau `map[string][]*websocket.Conn` dengan mutex
- Saat pembayaran sukses (webhook) → payment service memanggil `hub.Broadcast(streamKey, payload)`
- Koneksi yang putus (read error) → unregister otomatis dari hub
- Hub diinisialisasi sekali di `main.go` dan di-inject ke handler via dependency injection

### Payment — Midtrans

- Gunakan **Midtrans Snap API** untuk membuat token pembayaran
- Order ID format: `{username}-{unix_timestamp}-{random_6_char}`
- Webhook `/payment/webhook`:
    1. Verifikasi `signature_key` (SHA-512 dari `order_id + status_code + gross_amount + server_key`)
    2. Cek `transaction_status` dan `fraud_status`
    3. Jika sukses → update status donasi → hitung net amount (setelah platform fee) → tambah saldo streamer → push WebSocket event
- Jangan proses webhook yang duplikat (cek status donasi sudah `success` sebelum update)

### Environment Variables

```
PORT=8080
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable
JWT_SECRET=
JWT_REFRESH_SECRET=
JWT_EXPIRY_HOURS=1
JWT_REFRESH_EXPIRY_DAYS=7
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
MIDTRANS_SERVER_KEY=
MIDTRANS_CLIENT_KEY=
MIDTRANS_ENVIRONMENT=sandbox
PLATFORM_FEE_PERCENT=5
BASE_URL=https://yourdomain.com
```

## API Endpoints

### Auth

- `POST /auth/register` — email, username, password
- `POST /auth/login` — email, password → JWT token
- `GET  /auth/google` — redirect OAuth
- `GET  /auth/google/callback`
- `POST /auth/refresh` — refresh token

### User

- `GET  /users/me` _(JWT)_
- `PUT  /users/me` _(JWT)_ — update profil
- `GET  /users/:username` — profil publik

### Overlay _(semua JWT)_

- `GET  /overlay/settings`
- `PUT  /overlay/alert`
- `PUT  /overlay/template`
- `PUT  /overlay/filter`
- `PUT  /overlay/sound`
- `POST /overlay/stream-key/reset`

### Donasi

- `POST /donate/:username` — publik, submit donasi → return snap_token
- `GET  /donations` _(JWT)_ — histori masuk
- `GET  /donations/:id` _(JWT)_

### Payment

- `POST /payment/webhook` — Midtrans webhook (no auth, tapi verifikasi signature)

### Wallet _(semua JWT)_

- `GET  /wallet/balance`
- `POST /wallet/cashout`
- `GET  /wallet/cashout/history`

### WebSocket

- `WS /ws/overlay?key={streamKey}` — OBS browser source connection

## Keamanan

- Selalu validasi dan sanitasi input di handler layer
- Gunakan prepared statements / parameterized queries — **tidak boleh** string interpolation di SQL
- Rate limit endpoint publik (`/donate/:username`, `/payment/webhook`)
- Validasi URL media hanya izinkan domain: `youtube.com`, `youtu.be`, `tiktok.com`
- Stream key: simpan sebagai bcrypt hash di DB, bandingkan dengan `bcrypt.CompareHashAndPassword`
- Jangan log sensitive data (token, password, server key)
- Midtrans webhook: tolak request jika signature tidak cocok (return 400)

## Testing

- Unit test untuk `service.go` — mock repository dengan interface (tidak perlu library mock, cukup manual struct)
- Integration test handler menggunakan `net/http/httptest` + `gin.New()` (bukan `gin.Default()` agar tidak ada logger noise)
- Test WebSocket handler menggunakan `gorilla/websocket` dialer ke test server
- Test webhook handler dengan skenario: sukses, failed, duplikat, signature tidak valid

```go
// Contoh test Gin handler
func TestCreateHandler(t *testing.T) {
    mockService := &MockService{}
    handler := NewHandler(mockService)

    r := gin.New()
    r.POST("/resource", handler.Create)

    w := httptest.NewRecorder()
    body := strings.NewReader(`{"field": "value"}`)
    req := httptest.NewRequest(http.MethodPost, "/resource", body)
    req.Header.Set("Content-Type", "application/json")

    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusCreated, w.Code)
}
```
