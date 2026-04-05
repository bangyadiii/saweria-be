# TODO — saweria-be (Golang)

Legend: ✅ Selesai · ❌ Belum dibuat · 🔧 Sebagian

---

## 0. Setup Awal

- ✅ `go mod init` — inisialisasi Go module
- ✅ Setup struktur folder sesuai PRD (`cmd/`, `internal/`, `pkg/`, `migrations/`)
- ✅ Install dependency utama:
    - `github.com/gin-gonic/gin`
    - `github.com/golang-jwt/jwt/v5`
    - `github.com/jmoiron/sqlx` + `github.com/lib/pq`
    - `github.com/midtrans/midtrans-go` (Core API)
    - `github.com/gorilla/websocket`
    - `github.com/joho/godotenv`
    - `golang.org/x/crypto/bcrypt`
    - `golang.org/x/oauth2` + `golang.org/x/oauth2/google`
    - `github.com/gin-contrib/cors`
- ✅ `pkg/config/env.go` — load dan validasi semua env variable saat startup
- ✅ `pkg/database/postgres.go` — inisialisasi koneksi PostgreSQL, ping check
- ✅ `cmd/main.go` — entry point: setup DB, router, middleware global, jalankan server
- ✅ `.env.example` file dengan semua variable yang dibutuhkan

---

## 1. Database Migrations

- ✅ `migrations/001_create_users.up/down.sql`
- ✅ `migrations/002_create_overlay_settings.up/down.sql`
- ✅ `migrations/003_create_donations.up/down.sql` — kolom `payment_token` (bukan snap_token)
- ✅ `migrations/004_create_cashouts.up/down.sql`
- ✅ `migrations/005_rename_snap_token.up/down.sql` — migrasi kolom lama ke `payment_token`
- ✅ `migrations/006_create_alert_queue.up/down.sql` — server-side alert queue table

---

## 2. Auth (`internal/auth/`)

- ✅ `repository.go` — `FindByEmail`, `FindByUsername`, `Create`
- ✅ `service.go` — `Register`, `Login`, `ValidateToken`, `RefreshToken`
- ✅ `handler.go` — `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`
- ✅ `google.go` — `GET /auth/google`, `GET /auth/google/callback`
- ✅ `pkg/middleware/auth.go` — JWT middleware

---

## 3. User (`internal/user/`)

- ✅ `repository.go` — `FindByID`, `FindByUsername`, `Update`, `UpdateProfileImage`
- ✅ `service.go` — `GetMe`, `GetPublicProfile`, `UpdateProfile`, `UploadProfileImage`
- ✅ `handler.go` — `GET /users/me`, `PUT /users/me`, `PUT /users/me/image`, `GET /users/:username`

---

## 4. Overlay Settings (`internal/overlay/`)

- ✅ `repository.go` — `FindByUserID`, `Upsert`, `UpdateStreamKey`, `FindByStreamKey`
- ✅ `service.go` — `GetSettings`, `UpdateAlertRules`, `UpdateTemplate`, `UpdateFilterKata`, `UpdateSound`, `ResetStreamKey`
- ✅ `handler.go` — `GET /overlay/settings`, `PUT /overlay/alert`, `PUT /overlay/template`, `PUT /overlay/filter`, `PUT /overlay/sound`, `POST /overlay/stream-key/reset`

---

## 5. Donasi (`internal/donation/`)

- ✅ `repository.go` — `Create`, `FindByStreamerID`, `FindByID`, `FindByOrderID`, `UpdateStatus`
- ✅ `service.go` — `Submit` (Core API charge), `GetHistory`, `GetDetail`
- ✅ `handler.go` — `POST /donate/:username`, `GET /donations`, `GET /donations/:id`

---

## 6. Payment / Midtrans (`internal/payment/`)

> Menggunakan **Core API** (bukan Snap). Mendukung: bank_transfer (BCA/BNI/BRI/Permata), echannel (Mandiri), gopay, shopeepay.

- ✅ `midtrans.go` — `NewMidtransClient`, `CreateCharge`, `parseChargeResponse`
- ✅ `service.go` — `ProcessWebhook` (SHA-512 verify, idempoten, update status, AddBalance, AlertQueue.Enqueue)
- ✅ `handler.go` — `POST /payment/webhook`

---

## 7. Wallet & Cashout (`internal/wallet/`)

- ✅ `repository.go` — `GetBalance`, `CreateCashout`, `DeductBalance`, `GetCashoutHistory`
- ✅ `service.go` — `GetBalance`, `RequestCashout`, `GetCashoutHistory`
- ✅ `handler.go` — `GET /wallet/balance`, `POST /wallet/cashout`, `GET /wallet/cashout/history`

---

## 8. WebSocket (`internal/websocket/`)

- ✅ `hub.go` — `Hub` struct, `Register`, `Unregister`, `Broadcast` (dengan `SetWriteDeadline` + dead conn cleanup)
- ✅ `handler.go` — `GET /ws?key={streamKey}`, manual ping/pong (`pingInterval=5s`, `pongWait=10s`), done channel, LIFO defers

---

## 9. Alert Queue (`internal/alertqueue/`)

> Server-side queue — alert dikirim satu per satu per streamer, tidak ada race condition.

- ✅ `repository.go` — `Enqueue`, `ClaimNext` (atomic SQL), `MarkDone`, `GetPendingStreamerIDs`, `ResetStaleProcessing`
- ✅ `manager.go` — per-streamer worker goroutine, `RecoverPending` on startup

---

## 10. Widgets (`internal/widgets/`)

- ✅ `repository.go` — lookup by stream key, leaderboard query
- ✅ `handler.go` — `GET /widgets/info?streamKey=`, `GET /widgets/leaderboard?streamKey=`

---

## 11. Middleware & Utilitas (`pkg/`)

- ✅ `pkg/middleware/auth.go` — JWT auth middleware
- ❌ `pkg/middleware/ratelimit.go` — rate limiter untuk endpoint publik (`/donate/:username`, `/payment/webhook`)

---

## 12. Konfigurasi & DevOps

- ✅ `.env.example` — template semua env variable
- ✅ `Makefile` — `run`, `build`, `test`, `vet`, `migrate-up`, `migrate-down`, `migrate-reset`, `docker-build`, `docker-run`
- ✅ `Dockerfile` — multi-stage build (golang:1.25-alpine → alpine:3.21)
- ✅ `docker-compose.yml` (di root) — db + backend + frontend
- ✅ CORS setup — izinkan origin frontend
