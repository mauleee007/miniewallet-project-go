# Product Requirement Document (PRD) & Technical Specification
**Project Name:** Mini E-Wallet & Distributed Transfer System (MiniWallet)  
**Document Version:** 1.0.0 (Final & Approved)  
**Status:** Locked for Development  
**Target Architecture:** Microservices, Event-Driven, Distributed Locking  

---

## 1. Tujuan dan Objektif (Goals)

### Visi & Misi
Membangun sistem *core backend* dompet digital berskala *enterprise* yang mampu memproses transaksi finansial berkecepatan tinggi, dijamin aman dari kebocoran saldo (*double-spending*), dan tahan terhadap lonjakan trafik menggunakan arsitektur berbasis *event-driven*.

### Metrik Keberhasilan (Success Metrics)
| Metrik | Target | Keterangan |
|---|---|---|
| **Zero Double-Spending** | 0% Insiden | Tidak boleh ada saldo minus atau transfer ganda saat terjadi *spam klik* (*concurrent requests*). |
| **High Throughput** | < 50ms (Read) / < 100ms (Write) | Waktu respons API untuk cek saldo dan eksekusi transfer uang. |
| **ACID Compliance** | 100% Akurat | Mutasi saldo pengirim dan penerima harus selalu seimbang tanpa ada uang yang hilang atau menggantung. |
| **Asynchronous Speed** | < 500ms | Waktu keterlambatan (*latency*) dari uang terkirim hingga notifikasi tercatat di sistem log. |

### Target Pengguna
1. **Nasabah Ritel:** Pengguna aplikasi yang melakukan *top-up*, cek saldo, dan transfer antar-akun.
2. **Sistem Internal / Auditor:** Sistem yang membutuhkan catatan riwayat transaksi (*audit trail*) yang permanen dan tidak bisa dimanipulasi.

---

## 2. Ruang Lingkup dan Fitur (Scope & Features)

### Daftar Fitur & Prioritas (MoSCoW Method - MVP Phase)
| Fitur | Prioritas | Deskripsi Singkat |
|---|---|---|
| **Registrasi & Login (JWT)** | Must-Have | Otentikasi pengguna menggunakan token JWT yang divalidasi via Redis. |
| **Top-Up Saldo (Dummy)** | Must-Have | Menambah saldo dompet digital untuk keperluan pengujian transaksi. |
| **Transfer Antar-Akun** | Must-Have | Fitur inti memotong saldo pengirim dan menambah saldo penerima secara atomik. |
| **Distributed Lock Anti-Spam** | Must-Have | Mencegah *race condition* menggunakan penguncian (*mutex lock*) di Redis. |
| **Asynchronous Notification** | Must-Have | Mengirim log/notifikasi transaksi sukses di latar belakang menggunakan Apache Kafka. |
| **Riwayat Mutasi (Ledger)** | Should-Have | Menampilkan daftar uang masuk dan keluar beserta *timestamp*. |
| **Limit Transaksi Harian** | Nice-to-Have | Membatasi maksimal nominal transfer harian per akun (Fase 2). |

### Keterbatasan (Out of Scope / Non-Goals)
* **UI/UX Mobile App:** Fokus murni pada pengembangan RESTful API yang diuji via Postman/cURL.
* **Integrasi Payment Gateway Asli:** Fitur *top-up* menggunakan sistem *dummy inject* langsung di DB/API tanpa terhubung ke jaringan perbankan asli.
* **Fitur Kredit / Paylater:** Fokus di fase ini hanya pada perputaran uang tunai digital (*debit-based*).

---

## 3. Alur Pengguna dan Desain (User Experience & Flow)

### Alur Kerja Inti: Transfer Saldo (The Golden Flow)
1. **Request:** User mengirim API `POST /api/v1/transfers` berisi `recipient_id` and `amount` dengan *header* JWT.
2. **Auth Guard:** API Gateway memvalidasi token JWT ke **Redis**. Jika tidak sah, tolak (`401 Unauthorized`).
3. **Distributed Locking:** *Wallet Service* meminta kunci (*mutex lock*) ke **Redis** untuk `wallet_id` pengirim maksimal 3 detik.
   * *Jika gagal dikunci (sedang ada proses transfer lain):* Tolak (`429 Too Many Requests` / *System Busy*).
4. **Database Transaction:** *Wallet Service* membuka transaksi di **MySQL via GORM**:
   * Cek apakah saldo pengirim >= `amount`. Jika kurang, *Rollback* (`400 Bad Request`).
   * Kurangi saldo pengirim sebesar `amount`.
   * Tambah saldo penerima sebesar `amount`.
   * Catat rekam jejak ke tabel `transactions`.
   * *Commit* transaksi.
5. **Release Lock:** Kunci di **Redis** dilepaskan agar akun bisa bertransaksi kembali.
6. **Publish Event:** *Wallet Service* melempar event `TransactionCompleted` ke topik **Apache Kafka**.
7. **Instant Response:** API langsung membalas ke User: `200 OK - Transfer Berhasil` (tanpa menunggu proses notifikasi selesai).
8. **Background Consume:** **Notification Service** menyedot pesan dari Kafka, lalu mencatat log atau mengeksekusi simulasi pengiriman email bukti transfer.

### Wireframe Balikan JSON (API Contract Preview)
**Response Success (HTTP 200 OK):**
```json
{
  "status": "success",
  "message": "Transfer berhasil diproses",
  "data": {
    "transaction_id": "tx-882910-xzy",
    "sender_id": "usr-1123-abc",
    "recipient_id": "usr-9981-pqr",
    "amount": 50000,
    "current_balance": 150000,
    "created_at": "2026-07-18T21:05:00Z"
  }
}
```

---

## 4. Persyaratan Teknis & Lingkungan

### Kompatibilitas & Infrastruktur
* **Antarmuka Komunikasi:** RESTful API dengan format payload JSON standar.
* **Lingkungan Eksekusi:** Harus bisa dijalankan sepenuhnya di dalam **Docker Container** (Linux Alpine/Ubuntu based) secara terisolasi tanpa menginstal *tool* database secara manual di OS lokal.

### Standar Kinerja & Keamanan
* **Keamanan Konkurensi:** Tahan terhadap pengujian *Load Testing* (Apache Bench / k6) dengan 100 *request* transfer bersamaan pada 1 akun tanpa menyebabkan saldo minus (*zero race condition*).
* **Idempotency:** Setiap transaksi memiliki `reference_id` unik untuk mencegah pemrosesan ganda akibat gangguan jaringan.
* **Password Storage:** Wajib menggunakan algoritma *hashing* `bcrypt` dengan *cost* minimal 10.

---

## 5. Batasan, Asumsi, dan Ketergantungan

* **Asumsi:** Klien (Postman/Mobile) selalu menyertakan token JWT yang valid di *header* setelah login berhasil.
* **Batasan Infrastruktur:** Seluruh ekosistem (3 Service Go, MySQL, Redis, Kafka, Zookeeper) dijalankan menggunakan spesifikasi lokal minimal RAM 4GB dan 2 Core CPU via Docker Compose.
* **Ketergantungan:** 
  * *Wallet Service* bergantung pada **Redis**; jika Redis mati, transaksi transfer harus ditolak demi keamanan (*fail-safe*).
  * *Notification Service* bergantung pada **Kafka**; jika Kafka lambat, tidak boleh memengaruhi kecepatan *response time* API utama.

---

## 6. Spesifikasi Tech Stack (Locked & Sealed 🔒)

PENTING: Seluruh implementasi kode wajib mematuhi batasan teknologi di bawah ini tanpa pengecualian.

| Lapisan (Layer) | Teknologi / Library | Alasan Pemilihan & Penggunaan |
|---|---|---|
| **Bahasa Pemrograman** | **Golang 1.22+** | Eksekusi cepat, manajemen memori efisien, dan *Goroutine* untuk *concurrency* tinggi. |
| **Web Framework / Router** | **Chi v5** (`github.com/go-chi/chi/v5`) | Sangat ringan, modular, dan 100% kompatibel dengan `net/http` standar Go. |
| **Security & Auth** | **JWT v5 & Bcrypt** (`github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`) | Standar industri untuk otentikasi *stateless* token yang aman dan *hashing password* satu arah dengan *salt*. |
| **Relational Database** | **MySQL 8.0** | Standar industri perbankan yang stabil untuk menangani relasi data finansial. |
| **ORM & Database Tool** | **GORM v2** (`gorm.io/gorm`) | Pemodelan tabel (*Auto-Migration*), *Hooks*, dan pengelolaan *Database Transaction* (`tx.Begin()`). |
| **In-Memory Cache & Lock** | **Redis 7.0** (`github.com/redis/go-redis/v9`) | Caching sesi JWT dan eksekusi Distributed Mutex Lock untuk sistem anti *double-spending*. |
| **Message Broker** | **Apache Kafka** (`github.com/segmentio/kafka-go`) | Broker event-driven untuk memecah alur transfer yang *synchronous* menjadi *asynchronous*. |
| **Container & Orchestration** | **Docker & Docker Compose** | Mengisolasi seluruh *database*, *broker*, dan *microservices* ke dalam satu *Docker Network*. |

---

## 7. Desain Arsitektur Microservices

Sistem dipecah menjadi 3 layanan terisolasi yang berkomunikasi dalam jaringan Docker yang sama:
1. **`auth-service`**: Mengelola pengguna, otentikasi, pembuatan token JWT, dan validasi sesi di Redis.
2. **`wallet-service`**: *Core engine* yang menangani saldo, mutasi, penguncian Redis Lock, transaksi MySQL via GORM, dan *producer* event ke Kafka.
3. **`notification-service`**: *Background worker* yang bertindak sebagai *consumer* Kafka untuk mencatat log aktivitas dan pengiriman bukti transaksi.
