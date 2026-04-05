# saweria-be

Backend REST API untuk platform donasi streamer Saweria, dibangun dengan Go + Gin.

## Tech Stack

- **Go 1.25** — bahasa utama
- **Gin** — HTTP router & middleware
- **PostgreSQL 17** — database utama (via `sqlx` + `pq`)
- **JWT** — autentikasi (`golang-jwt/jwt/v5`)
- **Midtrans Core API** — payment gateway (bank transfer, GoPay, ShopeePay)
- **Gorilla WebSocket** — push notifikasi alert donasi real-time ke OBS
- **golang-migrate** — database migration

## Struktur Folder

```
saweria-be/
├── cmd/            # Entry point (main.go)
├── internal/
│   ├── auth/       # Register, login, JWT, Google OAuth
│   ├── user/       # Profil pengguna
│   ├── overlay/    # Pengaturan alert OBS, stream key
│   ├── donation/   # Submit & riwayat donasi
│   ├── payment/    # Midtrans webhook handler
│   ├── wallet/     # Saldo & cashout
│   ├── websocket/  # Hub & handler WebSocket
│   ├── alertqueue/ # Server-side alert queue per streamer
│   ├── widgets/    # Widget publik (info, leaderboard)
│   └── domain/     # Shared domain types
├── pkg/
│   ├── config/     # Load & validasi env variable
│   ├── database/   # Koneksi PostgreSQL
│   └── middleware/ # JWT auth middleware
├── migrations/     # SQL migration files (up/down)
├── Makefile
├── Dockerfile
└── .env.example
```

## Prasyarat

- Go 1.21+
- PostgreSQL 17
- [golang-migrate](https://github.com/golang-migrate/migrate) (opsional, bisa pakai `make migrate-up`)

## Setup Lokal

### 1. Clone & konfigurasi env

```bash
cp .env.example .env
# Edit .env sesuai konfigurasi lokal
```

### 2. Buat database PostgreSQL

```sql
CREATE USER saweria WITH PASSWORD 'saweria';
CREATE DATABASE saweria OWNER saweria;
```

### 3. Jalankan migrasi

```bash
make migrate-up
```

### 4. Jalankan server

```bash
make run
# Server berjalan di http://localhost:8080
```

## Environment Variables

| Variable               | Keterangan                           | Contoh                                                        |
| ---------------------- | ------------------------------------ | ------------------------------------------------------------- |
| `PORT`                 | Port server                          | `8080`                                                        |
| `DATABASE_URL`         | PostgreSQL connection string         | `postgres://user:pass@localhost:5432/saweria?sslmode=disable` |
| `JWT_SECRET`           | Secret untuk access token            | _(string panjang acak)_                                       |
| `JWT_REFRESH_SECRET`   | Secret untuk refresh token           | _(string panjang acak)_                                       |
| `GOOGLE_CLIENT_ID`     | Google OAuth client ID               | `xxx.apps.googleusercontent.com`                              |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret           | `GOCSPX-xxx`                                                  |
| `MIDTRANS_SERVER_KEY`  | Midtrans server key                  | `SB-Mid-server-xxx`                                           |
| `MIDTRANS_CLIENT_KEY`  | Midtrans client key                  | `SB-Mid-client-xxx`                                           |
| `MIDTRANS_ENVIRONMENT` | `sandbox` atau `production`          | `sandbox`                                                     |
| `PLATFORM_FEE_PERCENT` | Persentase fee platform              | `5`                                                           |
| `BASE_URL`             | URL backend (untuk callback)         | `http://localhost:8080`                                       |
| `FRONTEND_URL`         | URL frontend (untuk CORS & redirect) | `http://localhost:3000`                                       |

## API Endpoints

### Auth

| Method | Path                    | Keterangan               |
| ------ | ----------------------- | ------------------------ |
| `POST` | `/auth/register`        | Daftar akun baru         |
| `POST` | `/auth/login`           | Login, dapat JWT         |
| `POST` | `/auth/refresh`         | Refresh access token     |
| `GET`  | `/auth/google`          | Redirect ke Google OAuth |
| `GET`  | `/auth/google/callback` | Callback Google OAuth    |

### User

| Method | Path               | Auth | Keterangan          |
| ------ | ------------------ | ---- | ------------------- |
| `GET`  | `/users/me`        | ✅   | Data profil sendiri |
| `PUT`  | `/users/me`        | ✅   | Update profil       |
| `PUT`  | `/users/me/image`  | ✅   | Upload foto profil  |
| `GET`  | `/users/:username` | —    | Profil publik       |

### Overlay Settings

| Method | Path                        | Auth | Keterangan                     |
| ------ | --------------------------- | ---- | ------------------------------ |
| `GET`  | `/overlay/settings`         | ✅   | Ambil semua pengaturan overlay |
| `PUT`  | `/overlay/alert`            | ✅   | Update aturan alert            |
| `PUT`  | `/overlay/template`         | ✅   | Update template teks alert     |
| `PUT`  | `/overlay/filter`           | ✅   | Update filter kata             |
| `PUT`  | `/overlay/sound`            | ✅   | Upload sound alert             |
| `POST` | `/overlay/stream-key/reset` | ✅   | Generate stream key baru       |

### Donasi

| Method | Path                | Auth | Keterangan                     |
| ------ | ------------------- | ---- | ------------------------------ |
| `POST` | `/donate/:username` | —    | Kirim donasi (Core API charge) |
| `GET`  | `/donations`        | ✅   | Riwayat donasi masuk           |
| `GET`  | `/donations/:id`    | ✅   | Detail donasi                  |

### Payment

| Method | Path               | Keterangan                               |
| ------ | ------------------ | ---------------------------------------- |
| `POST` | `/payment/webhook` | Callback Midtrans (HMAC-SHA512 verified) |

### Wallet

| Method | Path                      | Auth | Keterangan        |
| ------ | ------------------------- | ---- | ----------------- |
| `GET`  | `/wallet/balance`         | ✅   | Saldo wallet      |
| `POST` | `/wallet/cashout`         | ✅   | Ajukan pencairan  |
| `GET`  | `/wallet/cashout/history` | ✅   | Riwayat pencairan |

### WebSocket

| Method | Path                  | Keterangan                          |
| ------ | --------------------- | ----------------------------------- |
| `GET`  | `/ws?key={streamKey}` | Koneksi WebSocket untuk OBS overlay |

### Widgets (Publik)

| Method | Path                              | Keterangan                  |
| ------ | --------------------------------- | --------------------------- |
| `GET`  | `/widgets/info?streamKey=`        | Info streamer by stream key |
| `GET`  | `/widgets/leaderboard?streamKey=` | Top donatur                 |

## Makefile Commands

```bash
make run            # Jalankan server lokal
make build          # Build binary (CGO_ENABLED=0)
make test           # Jalankan semua test
make vet            # go vet
make lint           # golangci-lint
make migrate-up     # Jalankan semua migrasi
make migrate-down   # Rollback 1 migrasi
make migrate-reset  # Drop semua tabel + migrate-up ulang
make docker-build   # Build Docker image
make docker-run     # Jalankan container
```

## Docker

Build & jalankan dengan Docker Compose dari root proyek:

```bash
# Dari folder root (d:\Development\saweria\)
docker compose up -d --build
```

Atau build image backend saja:

```bash
make docker-build
make docker-run
```

## WebSocket Protocol

Client (OBS browser source) terhubung ke `/ws?key={streamKey}`. Server mengirim pesan JSON saat ada donasi:

```json
{
    "type": "donation",
    "donor_name": "Budi",
    "amount": 50000,
    "message": "GG!",
    "media_url": "https://youtube.com/watch?v=..."
}
```

Koneksi dijaga dengan manual ping/pong (`pingInterval=5s`, `pongWait=10s`). Client yang tidak merespons pong akan didisconnect otomatis.

## Alert Queue

Alert donasi dikirim satu per satu per streamer menggunakan server-side queue yang disimpan di PostgreSQL. Setiap streamer memiliki goroutine worker tersendiri untuk mencegah race condition antar penonton yang terhubung bersamaan.
