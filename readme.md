# Anti Fraud Gateway

High-performance API Gateway untuk deteksi fraud secara real-time, dibangun dengan **Go (Fiber)**, **Redis**, dan **MongoDB**.

Gateway ini dirancang untuk menjadi lapisan pertahanan antara aplikasi mobile (SDK) dan layanan backend utama. Setiap request yang masuk akan melewati proses dekripsi, autentikasi, pembatasan laju, dan evaluasi aturan fraud secara paralel sebelum diteruskan.

---

## Arsitektur Sistem

```
┌─────────────────────────────────────────────────────────────────┐
│                      CLIENT / MOBILE SDK                        │
│              (mengirim payload terenkripsi AES-256-GCM)         │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                     ANTI-FRAUD GATEWAY                           │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐     │
│  │  Rate Limiter │→│  API Key Auth│ →│  AES-GCM Decrypt    │     │
│  │  (Redis)      │ │  (X-API-Key) │  │  (payload → JSON)   │     │
│  └──────────────┘  └──────────────┘  └─────────┬───────────┘     │
│                                                 │                │
│                                                 ▼                │
│                                  ┌──────────────────────────┐    │
│                                  │   FRAUD RULE ENGINE      │    │
│                                  │   (Parallel Goroutines)  │    │
│                                  │                          │    │
│                                  │  ├─ CheckBlacklist       │    │
│                                  │  ├─ CheckMockGPS         │    │
│                                  │  └─ CheckTimestamp       │    │
│                                  └────────────┬─────────────┘    │
│                                               │                  │
│                               ┌───────────────┼───────────┐      │
│                               ▼               ▼           │      │
│                          FRAUD ✗         CLEAN ✅        │      │
│                          403              200             │      │
│                               │               │           │      │
│                               └───────┬───────┘           │      │
│                                       ▼                   │      │
│                              ┌─────────────────┐          │      │
│                              │  Audit Logger   │          │      │
│                              │  (Async Channel)│          │      │
│                              └────────┬────────┘          │      │
│                                       ▼                   │      │
│                              ┌─────────────────┐          │      │
│                              │  Worker Pool    │          │      │
│                              │  (3 Goroutines) │          │      │
│                              └────────┬────────┘          │      │
│                                       ▼                   │      │
│                              ┌─────────────────┐          │      │
│                              │    MongoDB      │          │      │
│                              │  (audit_logs)   │          │      │ 
│                              └─────────────────┘          │      │
└──────────────────────────────────────────────────────────────────┘
```

---

## Tech Stack

| Teknologi | Fungsi |
|---|---|
| **Go 1.24** | Bahasa pemrograman utama |
| **Fiber v2** | Web framework (seperti Express.js untuk Go) |
| **Redis** | Rate limiter counter + device blacklist storage |
| **MongoDB** | Penyimpanan audit log |
| **AES-256-GCM** | Enkripsi/dekripsi payload end-to-end |
| **Goroutines** | Evaluasi aturan fraud secara paralel |
| **Channels** | Async audit logging (non-blocking) |

---

## Struktur Proyek

```
antiFraudGateway/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point, routing, inisialisasi koneksi
│
├── pkg/
│   ├── crypto/
│   │   └── aes.go                  # AES-256-GCM encrypt & decrypt
│   │
│   ├── middleware/
│   │   ├── auth.go                 # Middleware autentikasi API Key
│   │   └── ratelimiter.go          # Middleware rate limiter berbasis Redis
│   │
│   ├── fraud/
│   │   ├── model.go                # Struct DevicePayload & RuleResult
│   │   ├── rules.go                # Aturan pengecekan fraud
│   │   └── engine.go               # Orchestrator evaluasi paralel
│   │
│   └── audit/
│       ├── model.go                # Struct AuditLog
│       └── worker.go               # Worker pool untuk tulis ke MongoDB
│
├── .env                            # Konfigurasi environment
├── .air.toml                       # Konfigurasi hot-reload (development)
├── go.mod                          # Go module dependencies
└── go.sum                          # Checksum dependencies
```

---

## Konfigurasi Environment

Buat file `.env` di root project:

```env
PORT=3000
REDIS_URL=localhost:6379
REDIS_PASSWORD=
AES_SECRET_KEY=
API_KEY=
MONGODB_URI=
MONGODB_DATABASE=
```

| Variabel | Deskripsi |
|---|---|
| `PORT` | Port server (default: 3000) |
| `REDIS_URL` | Alamat Redis server |
| `REDIS_PASSWORD` | Password Redis (kosongkan jika tidak ada) |
| `AES_SECRET_KEY` | Kunci AES-256 (harus tepat 32 karakter) |
| `API_KEY` | API key untuk autentikasi request |
| `MONGODB_URI` | Connection string MongoDB |
| `MONGODB_DATABASE` | Nama database MongoDB |

---

## Cara Menjalankan

### Prasyarat

- Go 1.24+
- Redis (running di localhost:6379)
- MongoDB (running di 127.0.0.1:27017)

### Instalasi

```bash
# Clone repository
git clone https://github.com/your-username/antiFraudGateway.git
cd antiFraudGateway

# Install dependencies
go mod tidy

# Jalankan server
go run ./cmd/api/
```

### Dengan Hot-Reload (Development)

```bash
# Install Air
go install github.com/air-verse/air@latest

# Jalankan dengan auto-reload
air
```

Server akan berjalan di `http://localhost:3000`

---

## API Endpoints

### Route Publik (Tanpa Autentikasi)

| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/` | Welcome message |
| `GET` | `/health` | Health check |

### Route Terproteksi (Wajib Header `X-API-Key`)

Semua endpoint di bawah memerlukan header:
```
X-API-Key: <your-api-key>
```

| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/api/v1/testEncrypt` | Mendapatkan payload dummy terenkripsi |
| `POST` | `/api/v1/testEncrypt` | Mengenkripsi body JSON kustom |
| `POST` | `/api/v1/ingest` | **Endpoint utama** — terima, dekripsi, evaluasi fraud |
| `POST` | `/api/v1/blacklist` | Menambahkan device ke blacklist |
| `DELETE` | `/api/v1/blacklist` | Menghapus device dari blacklist |
| `GET` | `/api/v1/logs` | Melihat seluruh audit log |

---

## Dokumentasi API Detail

### 1. `GET /api/v1/testEncrypt`

Menghasilkan payload dummy terenkripsi dengan timestamp saat ini. Berguna untuk testing.

**Response:**
```json
{
  "payload": "base64-encrypted-string...",
  "raw_debug": "{\"device_id\":\"DEV-999\",\"latitude\":-5.3971,...}"
}
```

---

### 2. `POST /api/v1/testEncrypt`

Mengenkripsi body JSON kustom yang dikirimkan. Berguna untuk membuat skenario test fraud.

**Request Body:**
```json
{
  "device_id": "DEV-TEST",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "is_mock_location": true,
  "timestamp": 1000000000
}
```

**Response:**
```json
{
  "payload": "base64-encrypted-string...",
  "raw_debug": "..."
}
```

---

### 3. `POST /api/v1/ingest` Endpoint Utama

Menerima payload terenkripsi, mendekripsi, menjalankan evaluasi fraud, dan mencatat audit log.

**Request Body:**
```json
{
  "payload": "base64-encrypted-string-dari-testEncrypt"
}
```

**Response Sukses (Clean) — `200 OK`:**
```json
{
  "status": "clean",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Data lolos semua pengecekan fraud",
  "payload": {
    "device_id": "DEV-999",
    "latitude": -5.3971,
    "longitude": 105.2668,
    "is_mock_location": false,
    "timestamp": 1745190000
  },
  "all_checks": [
    { "rule_name": "device_blacklist", "is_fraud": false },
    { "rule_name": "mock_gps", "is_fraud": false },
    { "rule_name": "abnormal_timestamp", "is_fraud": false }
  ]
}
```

**Response Fraud Terdeteksi — `403 Forbidden`:**
```json
{
  "status": "fraud_detected",
  "request_id": "a1b2c3d4-...",
  "message": "Request ditolak karena terdeteksi aktivitas mencurigakan",
  "violations": [
    {
      "rule_name": "mock_gps",
      "is_fraud": true,
      "reason": "Terdeteksi penggunaan lokasi palsu (Mock GPS)"
    }
  ],
  "all_checks": [...]
}
```

---

### 4. `POST /api/v1/blacklist`

Menambahkan device ke daftar hitam.

**Request Body:**
```json
{
  "device_id": "DEV-HACKER"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Device 'DEV-HACKER' berhasil ditambahkan ke blacklist"
}
```

---

### 5. `DELETE /api/v1/blacklist`

Menghapus device dari daftar hitam.

**Request Body:**
```json
{
  "device_id": "DEV-HACKER"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Device 'DEV-HACKER' berhasil dihapus dari blacklist"
}
```

---

### 6. `GET /api/v1/logs`

Mengambil semua audit log yang tersimpan di MongoDB.

**Response:**
```json
{
  "total": 2,
  "logs": [
    {
      "request_id": "550e8400-...",
      "device_id": "DEV-999",
      "ip": "127.0.0.1",
      "endpoint": "/api/v1/ingest",
      "status": "clean",
      "payload": {...},
      "created_at": "2026-04-21T00:15:00Z"
    },
    {
      "request_id": "a1b2c3d4-...",
      "device_id": "DEV-HACKER",
      "ip": "127.0.0.1",
      "endpoint": "/api/v1/ingest",
      "status": "fraud_detected",
      "violations": [
        "Terdeteksi penggunaan lokasi palsu (Mock GPS)"
      ],
      "payload": {...},
      "created_at": "2026-04-21T00:16:00Z"
    }
  ]
}
```

---

## Aturan Fraud (Rule Engine)

Semua aturan dijalankan **secara paralel** menggunakan Goroutines + WaitGroup:

| Rule | Pengecekan | Aksi Jika Gagal |
|---|---|---|
| **Device Blacklist** | Apakah `device_id` ada di Redis SET `blacklist:devices`? | Fraud — device sudah di-blacklist |
| **Mock GPS** | Apakah `is_mock_location` bernilai `true`? | Fraud — lokasi palsu terdeteksi |
| **Abnormal Timestamp** | Apakah `timestamp` > 5 menit lalu (replay attack) atau dari masa depan? | Fraud — potensi manipulasi waktu |

---

## Lapisan Keamanan

| Layer | Mekanisme | HTTP Status |
|---|---|---|
| **Rate Limiter** | Maks 10 request/detik per IP (Redis counter) | `429 Too Many Requests` |
| **Autentikasi** | Header `X-API-Key` wajib dan valid | `401 Unauthorized` |
| **Enkripsi** | Payload dienkripsi AES-256-GCM end-to-end | `403 Forbidden` |
| **Fraud Detection** | 3 aturan paralel (blacklist, GPS, timestamp) | `403 Forbidden` |

---

## Audit Logging

Setiap request ke `/api/v1/ingest` (baik `clean` maupun `fraud_detected`) dicatat secara **asinkron** ke MongoDB:

- **Non-blocking** — audit log dikirim ke buffered channel (kapasitas 1000), tidak memperlambat response ke client
- **Worker Pool** — 3 goroutine background yang membaca dari channel dan menulis ke MongoDB
- **Persistent** — data tersimpan di collection `audit_logs` di database `antiFraudGateway`
- **Viewable** — dapat dilihat via `GET /api/v1/logs` atau langsung di **MongoDB Compass**

---

## Testing dengan Postman

### Persiapan

1. Pastikan Redis dan MongoDB sudah running
2. Jalankan server: `go run ./cmd/api/`
3. Set header global di Postman: `X-API-Key: <your-api-key>`

### Skenario Test

| # | Skenario | Langkah | Expected |
|---|---|---|---|
| 1 | Data bersih | GET testEncrypt → copy payload → POST ingest | `200` status `clean` |
| 2 | Mock GPS | POST testEncrypt dengan `is_mock_location: true` → POST ingest | `403` fraud `mock_gps` |
| 3 | Timestamp kadaluarsa | POST testEncrypt dengan `timestamp: 1000000000` → POST ingest | `403` fraud `abnormal_timestamp` |
| 4 | Blacklist device | POST blacklist → POST ingest | `403` fraud `device_blacklist` |
| 5 | Hapus blacklist | DELETE blacklist | `200` success |
| 6 | Multiple fraud | Kombinasi mock GPS + timestamp lama + blacklist → POST ingest | `403` dengan 3 violations |
| 7 | Rate limit | Spam 15 request dalam 1 detik | Request ke-11+ → `429` |
| 8 | Tanpa API Key | Request tanpa header X-API-Key | `401` |
| 9 | Cek audit log | GET /logs setelah beberapa test | Semua log tercatat |

---

## Catatan Pengembangan

Proyek ini dikembangkan dalam 5 fase:

| Fase | Deskripsi | Komponen Go |
|---|---|---|
| 1 | Setup server Fiber + koneksi Redis | `fiber.New()`, `redis.NewClient()` |
| 2 | Enkripsi AES-256-GCM end-to-end | `crypto/aes`, `crypto/cipher` |
| 3 | Middleware autentikasi + rate limiter | `fiber.Handler`, Redis `INCR` |
| 4 | Concurrent fraud rule engine | `goroutines`, `sync.WaitGroup`, `chan` |
| 5 | Async audit logging ke MongoDB | Worker pool, buffered `chan`, MongoDB driver |
