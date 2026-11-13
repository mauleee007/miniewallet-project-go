# 📘 MINI E-WALLET - PROJECT GUIDEBOOK & ARTIFACT JOURNAL

Buku panduan ini mendokumentasikan setiap langkah pengembangan sistem **Mini E-Wallet & Distributed Transfer System** secara bertahap. Dokumen ini mencatat keputusan arsitektural, implementasi kode, tujuan teknis, dan kepatuhan terhadap **Product Requirement Document (PRD)**. Gambaran visual sistem secara utuh (arsitektur, alur fitur, peran stack) ada di **[ConceptOfProject.md](ConceptOfProject.md)**.

---

## 📑 DAFTAR ISI
1. [Standar & Konvensi Arsitektur](#standar--konvensi-arsitektur)
2. [Modul 1: Auth Service (Identity & Authentication)](#modul-1-auth-service)
   - [Langkah 1: Setup Dependensi](#langkah-1-setup-dependensi)
   - [Langkah 2: Standarisasi Response JSON](#langkah-2-standarisasi-response-json)
   - [Langkah 3: Konfigurasi Terpusat (12-Factor App)](#langkah-3-konfigurasi-terpusat)
   - [Langkah 4: Koneksi Database & Connection Pooling](#langkah-4-koneksi-database--connection-pooling)
   - [Langkah 5: Domain Model, Interface & DTO](#langkah-5-domain-model-interface--dto)
   - [Langkah 6: Keamanan (Bcrypt & JWT) & Repository](#langkah-6-keamanan-bcrypt--jwt--repository)
   - [Langkah 7: Business Logic (Usecase Layer)](#langkah-7-business-logic-usecase-layer)
   - [Langkah 8: HTTP Delivery (Handler & Router)](#langkah-8-http-delivery-handler--router)
   - [Langkah 9: Wiring & Inisialisasi Aplikasi](#langkah-9-wiring--inisialisasi-aplikasi-entrypoint)
3. [Modul 2: Wallet Service (Core Engine & Balance)](#modul-2-wallet-service)
   - [Langkah 1: Setup Dependensi](#langkah-1-setup-dependensi-wallet-service)
   - [Langkah 2: Setup Fondasi (Config, DB, Response)](#langkah-2-setup-fondasi-wallet-service)
   - [Langkah 3: Domain Model (Wallet & Transaction Ledger)](#langkah-3-domain-model-wallet--transaction)
   - [Langkah 4: Repository Layer GORM](#langkah-4-repository-layer-gorm)
   - [Langkah 5: Usecase Top-Up & Cek Saldo](#langkah-5-usecase-top-up--cek-saldo)
   - [Langkah 6: HTTP Handler, Middleware, & Main](#langkah-6-http-handler-middleware--main-entrypoint)
   - [Langkah 7: Pola Transactor (Perbaikan R1)](#langkah-7-pola-transactor-perbaikan-r1--acid)
4. [Modul 3: Distributed Locking & Concurrency Guard](#modul-3-distributed-locking) *(Akan Datang)*
5. [Modul 4: Event-Driven Kafka Pipeline](#modul-4-event-driven-kafka) *(Akan Datang)*

---

## 🏛️ Standar & Konvensi Arsitektur

### Clean Architecture (Layered Hexagonal)
Setiap microservice diisolasi dengan struktur layer berikut:
- **`domain/`**: Entitas bisnis murni dan interface. Tidak boleh bergantung pada library eksternal atau database driver (*inward dependency rule*).
- **`repository/`**: Mengimplementasikan interface repository dari domain (berhubungan langsung dengan database, query SQL, Redis).
- **`usecase/`**: Menjalankan alur bisnis (*business rules*). Hanya berinteraksi dengan interface domain.
- **`delivery/`**: Layer transportasi data (HTTP Chi router, controller/handler, parsing payload).
- **`pkg/`**: Modul pembantu umum (*shared utilities*) yang dapat digunakan lintas layer.

### Aturan Komentar Kode vs Guidebook (Regulasi Proyek)
- **Di Dalam Kode Sumber:** Komentar HANYA berisi penomoran urutan proses yang ringkas, padat, dan jelas (contoh: `// 1. Validasi input`, `// 2. Cek duplikasi`).
- **Di Dalam `GUIDEBOOK.md`:** Seluruh rincian latar belakang arsitektur, parameter teknis, analisis keamanan, dan keterkaitan dengan PRD dicatat secara komprehensif di sini.

---

## 🔐 Modul 1: Auth Service

### Langkah 1: Setup Dependensi
* **Tujuan PRD:** Memenuhi Bab 6 (Spesifikasi Tech Stack - Locked & Sealed).
* **Rincian Library:**
  - `gorm.io/gorm` & `gorm.io/driver/mysql`: Database ORM untuk MySQL 8.0.
  - `github.com/redis/go-redis/v9`: In-memory cache dan validasi sesi.
  - `github.com/golang-jwt/jwt/v5`: Otentikasi stateless dengan token bertanda tangan kriptografis HS256.
  - `golang.org/x/crypto/bcrypt`: Pengacakan kata sandi satu arah (*one-way hash*) dengan salt.
  - `github.com/google/uuid`: Generator string UUID v4 acak 128-bit untuk ID pengguna.

---

### Langkah 2: Standarisasi Response JSON
* **File:** `services/auth/pkg/response/response.go`
* **Tujuan & Fungsi:**
  - Memastikan seluruh endpoint mengembalikan kontrak JSON yang seragam (PRD Bab 3).
  - `JSON(w, code, message, data)`: Mengirim HTTP status code, menyetel header `Content-Type: application/json`, membungkus payload ke dalam `SuccessResponse`, dan melakukan stream encoding via `json.NewEncoder(w).Encode(resp)`.
  - `Error(w, code, message)`: Menyeragamkan format error ke dalam `ErrorResponse`.
* **Kaidah Rekayasa:** Menggunakan `json.NewEncoder(w)` lebih hemat alokasi memori dibandingkan `json.Marshal()` karena data langsung dialirkan ke buffer socket jaringan tanpa membuat salinan slice byte sementara di memori Go.

---

### Langkah 3: Konfigurasi Terpusat (12-Factor App)
* **File:** `services/auth/internal/config/config.go`
* **Tujuan & Fungsi:**
  - Membaca konfigurasi dari *environment variable*: `PORT`, `DB_DSN`, `REDIS_ADDR`, dan `JWT_SECRET`.
  - Mengimplementasikan fungsi `getEnv(key, fallback)`: Memastikan service tidak crash saat booting di lingkungan lokal apabila environment tertentu belum diatur.

---

### Langkah 4: Koneksi Database & Connection Pooling
* **File:** `services/auth/internal/database/mysql.go` & `redis.go`
* **Tujuan & Analisis Teknis:**
  - **MySQL (GORM):**
    - `SetMaxIdleConns(10)`: Menjaga 10 koneksi tetap terbuka (*idle pool*), sehingga request yang datang mendadak tidak perlu membayar biaya latensi *three-way handshake* TCP berulang kali.
    - `SetMaxOpenConns(100)`: Membatasi maksimal 100 koneksi simultan untuk mencegah kehabisan soket MySQL (*Too many connections error*).
    - `SetConnMaxLifetime(time.Hour)`: Mendaur ulang koneksi setiap 1 jam untuk mencegah *stale socket* akibat firewall atau timeout NAT jaringan.
  - **Redis Client:**
    - Menerapkan *Fail-Fast Principle* dengan `context.WithTimeout(..., 5*time.Second)` saat menjalankan perintah `Ping()`. Jika server Redis mati, service langsung menolak menyala daripada membiarkan request klien menggantung (*hanging*).

---

### Langkah 5: Domain Model, Interface & DTO
* **File:** `services/auth/internal/domain/user.go`
* **Tujuan & Analisis Keamanan:**
  - **Entity `User`:** Representasi tabel database. Tag `json:"-"` pada `PasswordHash` menjamin secara kriptografis bahwa hash kata sandi tidak akan pernah bocor ke respons HTTP meskipun model di-serialize secara utuh.
  - **DTO (`RegisterRequest`, `LoginRequest`, `UserResponse`, `AuthResponse`):** Memisahkan data masukan dari klien dengan entitas database. Ini adalah mitigasi wajib terhadap kerentanan *Mass Assignment / Over-posting Vulnerability*.
  - **Interfaces (`UserRepository`, `AuthUsecase`):** Abstraksi kontrak murni. Memungkinkan pengujian unit (*unit testing*) menggunakan mock tanpa membutuhkan instance database fisik.

---

### Langkah 6: Keamanan (Bcrypt & JWT) & Repository
* **File:** `services/auth/pkg/utils/password.go`, `pkg/utils/token.go`, `internal/repository/mysql/user_repository.go`
* **Tujuan & Analisis Teknis:**
  - **Bcrypt (Cost 10):** Sesuai PRD baris 94 (*"Wajib menggunakan algoritma hashing bcrypt dengan cost minimal 10"*). Bcrypt secara otomatis menambahkan *salt* acak sehingga aman dari serangan *Rainbow Table*. Fungsi `CheckPasswordHash` mengevaluasi kecocokan dalam *constant-time* untuk mencegah serangan *Timing Attack*.
  - **JWT (v5 HS256):** Token berisi klaim standar (`Subject: user_id`, `ExpiresAt: 24 jam`, `IssuedAt`) dan klaim privat (`email`). Ditandatangani menggunakan algoritma simetris HMAC-SHA256 (`jwtSecret`).
  - **MySQL Repository:** Mengimplementasikan interface `domain.UserRepository`. Mengonversi error `gorm.ErrRecordNotFound` menjadi `nil, nil` agar layer usecase dapat mengontrol logika bisnis secara bersih tanpa ketergantungan pada tipe error database tertentu.

---

### Langkah 7: Business Logic (Usecase Layer)
* **File:** `services/auth/internal/usecase/auth_usecase.go`
* **Alur Proses & Analisis Keamanan:**
  1. **Registrasi:**
     - Validasi string kosong pada username, email, dan password.
     - Verifikasi panjang kata sandi minimal 6 karakter.
     - Pengecekan duplikasi email ke database (`FindByEmail`).
     - Hashing kata sandi via Bcrypt Cost 10.
     - Pembuatan UUID unik untuk User ID.
     - Penyimpanan ke database MySQL.
     - Pembuatan token JWT dan penyimpanan sesi aktif ke Redis (`session:user:<id>`) selama 24 jam (sesuai PRD baris 33 & 128).
  2. **Login:**
     - Menggunakan pesan kesalahan umum (*Generic Error Message*): `"kombinasi email atau password salah"`. Hal ini mencegah penyerang mengetahui apakah sebuah email sudah terdaftar atau belum (*User Enumeration Attack*).
     - Pembaruan sesi token di Redis saat login berhasil.

---

### Langkah 8: HTTP Delivery (Handler & Router)
* **File:** `services/auth/internal/delivery/http/auth_handler.go`
* **Tujuan & Fungsi:**
  - Mendaftarkan endpoint `POST /register` dan `POST /login` ke Chi router.
  - Membaca body request JSON via `json.NewDecoder(r.Body).Decode(&req)`.
  - Meneruskan data ke layer `AuthUsecase`.
  - Memetakan hasil bisnis ke HTTP status code yang tepat (`201 Created` untuk pendaftaran sukses, `200 OK` untuk login sukses, `400 Bad Request` untuk input tidak valid, `401 Unauthorized` untuk kredensial salah).

---

### Langkah 9: Wiring & Inisialisasi Aplikasi (Entrypoint)
* **File:** `services/auth/cmd/main.go`
* **Tujuan & Siklus Hidup Aplikasi:**
  - Titik masuk (*entry point*) aplikasi.
  - Memuat konfigurasi, menghubungkan MySQL dan Redis.
  - Menjalankan **GORM AutoMigrate** untuk entitas `domain.User` sehingga tabel otomatis terbuat saat booting pertama kali.
  - Merangkai *Dependency Injection* dari lapis terdalam hingga terluar: `DB -> Repository -> Usecase -> Handler -> Router`.
  - Memasang middleware global: `middleware.Logger` (pencatatan log akses HTTP) dan `middleware.Recoverer` (menangkap *runtime panic* agar proses server tidak mati mendadak).
  - Membuka port server HTTP via `http.ListenAndServe`.

---
*(Modul 1 Auth Service: SELESAI & TERVERIFIKASI ✅)*

---

## 💰 Modul 2: Wallet Service (Core Engine & Balance)

Modul ini bertanggung jawab atas integritas saldo dompet digital, mutasi transaksi (*ledger*), serta mekanisme anti *double-spending* menggunakan *Distributed Mutex Locking* di Redis.

### Langkah 1: Setup Dependensi Wallet Service
* **Library Eksternal:**
  - `gorm.io/gorm` & `driver/mysql`: Database engine untuk `wallet_db`.
  - `github.com/redis/go-redis/v9`: Client Redis untuk penguncian terdistribusi (*distributed lock*).
  - `github.com/golang-jwt/jwt/v5`: Pustaka parsing dan verifikasi token JWT pada middleware.
  - `github.com/segmentio/kafka-go`: Producer & Consumer pesan Kafka untuk pemrosesan event asinkron.
  - `github.com/google/uuid`: Generator identitas unik untuk Wallet ID, Transaction ID, dan Reference ID.

---

### Langkah 2: Setup Fondasi Wallet Service
* **File yang Dibuat:**
  1. `services/wallet/pkg/response/response.go`
  2. `services/wallet/internal/config/config.go`
  3. `services/wallet/internal/database/mysql.go`
  4. `services/wallet/internal/database/redis.go`
* **Tujuan & Analisis Teknis:**
  - **Separation of Database:** `wallet-service` memiliki database MySQL terisolasi (`wallet_db`) yang terpisah dari `auth_db` sesuai prinsip *Database-per-Service* pada arsitektur Microservices.
  - **Konfigurasi Tambahan:** Menyertakan pembacaan `KAFKA_BROKERS` untuk persiapan event streaming.

---

### Langkah 3: Domain Model (Wallet & Transaction)
* **File:** `services/wallet/internal/domain/wallet.go` & `transaction.go`
* **Analisis Rekayasa Finansial (Sesuai Regulasi PRD Keamanan Tinggi):**
  1. **Mengapa Saldo Menggunakan `int64` (Bukan `float64`):**
     - Angka desimal pada komputer direpresentasikan menggunakan standar biner IEEE 754. Operasi seperti `0.1 + 0.2` pada tipe `float64` menghasilkan `0.30000000000000004`. Dalam sistem perbankan/finansial, akumulasi selisih desimal ini dapat menyebabkan ketidakseimbangan neraca keuangan (*loss of cents*).
     - Menggunakan `int64` dengan satuan terkecil (Rupiah utuh / sen) menjamin akurasi mutlak 100% tanpa adanya kesalahan pembulatan.
  2. **Struktur Entitas yang Dibuat:**
     - `Wallet`: satu dompet per user (`uniqueIndex` pada `UserID`), mata uang default `IDR`.
     - `Transaction`: `ReferenceID` diberi `uniqueIndex` sebagai wadah *Idempotency-Key*; `SenderWalletID` bertipe `*string` (nullable) karena transaksi `TOPUP` tidak memiliki pengirim.
     - Enum `TransactionType` (`TOPUP`, `TRANSFER`) dan `TransactionStatus` (`SUCCESS`, `FAILED`) didefinisikan sebagai tipe string agar tidak ada nilai liar (*magic string*).
  3. **Rujukan Requirement:** Aturan desain sistem yang menjadi dasar entitas ini (ACID, Ledger/Audit Trail, Idempotency, Security PIN & HTTPS) dicatat di **PRD Bab 8 — Persyaratan Desain Sistem**, bukan di Guidebook, karena itu adalah *requirement*, bukan rekam jejak implementasi.

---

### Langkah 4: Repository Layer GORM (Wallet & Transaction)
* **File:** `services/wallet/internal/repository/mysql/wallet_repository.go` & `transaction_repository.go`
* **Tujuan & Analisis Rekayasa:**
  - **Wallet Repository:** Menyediakan akses ke entitas dompet. Fungsi `UpdateBalance` menerima opsi transaksi `tx *gorm.DB` untuk menjaga *ACID compliance*. Menggunakan query `gorm.Expr("balance + ?", amount)` untuk mencegah *race condition* dasar di level database.
  - **Transaction Repository:** Mencatat setiap aktivitas uang ke buku besar (*ledger*). Menerima transaksi database yang sama dengan Wallet Repository. Pengecekan *Idempotency Key* (`ReferenceID`) terjadi di sini lewat batasan indeks unik.

---

### Langkah 5: Usecase Top-Up & Cek Saldo
* **File:** `services/wallet/internal/usecase/wallet_usecase.go`
* **Tujuan & Analisis Rekayasa:**
  - **Lazy Initialization (`GetBalance`):** Jika saldo dicari dan dompet belum ada, Usecase memanggil `repo.Create` untuk membuatkan dompet kosong bersaldo 0.
  - **ACID Transaction (`TopUp`):** Menggunakan fitur transaksi GORM `db.Transaction(...)`. Di dalamnya, saldo ditambah dan ledger dicatat bersamaan. Jika salah satu operasi gagal, GORM melakukan *rollback* otomatis sehingga tidak ada anomali uang tercipta tanpa riwayat. Karena fitur ini *dummy*, `ReferenceID` di-generate via UUID di dalam Usecase.

---

### Langkah 6: HTTP Handler, Middleware, & Main (Entrypoint)
* **File:** 
  - `services/wallet/pkg/middleware/auth.go`
  - `services/wallet/internal/delivery/http/wallet_handler.go`
  - `services/wallet/cmd/main.go`
* **Tujuan & Analisis Rekayasa:**
  - **Auth Middleware:** Sesuai PRD, seluruh transaksi harus terlindungi. Middleware ini bertugas memotong (*intercept*) *request*, membaca token JWT dari *header* `Authorization`, memverifikasi *signature* menggunakan `JWT_SECRET`, dan mengekstrak `user_id` ke dalam *context*.
  - **Wallet Handler:** Mengubah *request* HTTP menjadi panggilan Usecase. Menggunakan helper `response.go` untuk standarisasi balikan JSON (sukses maupun error). Karena `TopUp` tidak butuh *body* kompleks, kita hanya mengekstrak `amount`.
  - **Main (Dependency Injection):** `main.go` merangkai semua komponen (*Database -> Repository -> Usecase -> Handler -> Router*) dan menjalankan *web server*. Ini adalah prinsip dasar injeksi dependensi pada *Clean Architecture*.

---

### Langkah 7: Pola Transactor (Perbaikan R1 — ACID)
* **File:**
  - `services/wallet/internal/domain/transactor.go` (baru)
  - `services/wallet/internal/repository/mysql/transactor.go` (baru)
  - `services/wallet/internal/repository/mysql/wallet_repository.go` & `transaction_repository.go`
  - `services/wallet/internal/usecase/wallet_usecase.go`
  - `services/wallet/cmd/main.go`
* **Tujuan & Kegunaan:** Memperbaiki bug R1. Sebelumnya usecase menyimpan `tx` dengan key `txKey("tx")` sedangkan repository membaca key `"tx"`. Di Go, key context dibandingkan berdasarkan **tipe + nilai**, sehingga keduanya dianggap berbeda → repository tidak pernah menemukan `tx` dan query berjalan di luar transaksi.
* **Rujukan PRD:** Bab 8 poin 1 — seluruh repository dalam satu alur bisnis wajib memakai objek transaksi yang sama (ACID Compliance).
* **Bedah Alur Kerja:**
  1. `domain.Transactor` adalah interface dengan satu method `WithinTransaction(ctx, fn)`. Usecase hanya mengenal interface ini, sehingga tidak lagi meng-import `gorm` (*inward dependency rule*).
  2. `mysql.transactor` membuka `db.Transaction(...)`, menyisipkan `tx` ke context dengan key `txKey{}`, lalu menjalankan `fn`. Jika `fn` mengembalikan error → GORM `ROLLBACK`; jika `nil` → `COMMIT`.
  3. `dbFromContext(ctx, fallback)` dipakai oleh semua method repository: kembalikan `tx` jika ada di context, jika tidak gunakan koneksi biasa.
  4. Key context (`txKey{}`) dan pembacanya berada di **satu package yang sama**, sehingga mustahil terjadi ketidakcocokan key lagi.
* **Catatan Rekayasa:**
  - `txKey` bertipe `struct{}` yang tidak di-export: package lain tidak bisa membuat key yang sama (anti-tabrakan) dan ukurannya 0 byte.
  - Di dalam closure wajib memakai `txCtx`, bukan `ctx` luar. Memakai `ctx` luar akan membuat query lari ke luar transaksi tanpa error.
  - `GetByID`/`GetByUserID` juga memakai `dbFromContext` agar nanti bisa dipakai untuk `SELECT ... FOR UPDATE` di fitur Transfer.
* **Uji Rollback (2026-11-13) — LULUS ✅:**
  1. Top-up normal 50.000 → saldo 50.000, ledger 1 baris.
  2. Pasang trigger MySQL yang menggagalkan setiap `INSERT` ke `transactions` (*fault injection* tanpa mengubah kode):
     ```sql
     CREATE TRIGGER wallet_db.tg_rollback_test BEFORE INSERT ON wallet_db.transactions
     FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'ROLLBACK TEST';
     ```
  3. Top-up 30.000 → log GORM menunjukkan `UPDATE wallets SET balance=balance + 30000` sempat dieksekusi, lalu `INSERT INTO transactions` gagal (Error 1644) → API membalas `500`.
  4. Hasil: saldo **tetap 50.000**, ledger **tetap 1 baris** → update saldo ikut di-*rollback*. ACID terbukti.
  5. Bersihkan: `DROP TRIGGER wallet_db.tg_rollback_test;`
  - **Pelajaran:** trigger `SIGNAL` adalah cara aman menyimulasikan kegagalan database tanpa menyentuh kode aplikasi, sehingga tidak ada risiko lupa mengembalikan kode uji.

---

## 🔍 Catatan Review Kode (2026-10-18)

Hasil review Modul 1 & 2 sebelum lanjut ke Modul 3. Diperbaiki satu per satu oleh pengembang; status diperbarui di tabel ini.

| # | Prioritas | Lokasi | Temuan | Status |
|---|---|---|---|---|
| R1 | 🔴 Kritis | `wallet/internal/usecase/wallet_usecase.go` + repository | Key context transaksi tidak cocok (`txKey("tx")` vs `"tx"`), sehingga repository tidak pernah memakai `tx` → update saldo & ledger berjalan **di luar** transaksi (melanggar ACID). | ✅ Selesai (Langkah 7, uji rollback lulus) |
| R2 | 🔴 Kritis | `wallet/pkg/middleware/auth.go` | `TrimPrefix("Bearer")` menyisakan spasi di depan token → semua token valid ditolak `401`. Algoritma JWT juga belum dikunci ke HS256. | Belum |
| R3 | 🔴 Kritis | `wallet/pkg/response/response.go` | Struct tag salah format (`json: "status"`, `amitempty`) → key JSON jadi `Status`/`Message`/`Data`, beda dengan kontrak PRD. | Belum |
| R4 | 🔴 Kritis | `Dockerfile` (auth & wallet) | Image `golang:1.22-alpine` tidak bisa build modul yang meminta `go 1.26.0` / `1.25.5` (image resmi memakai `GOTOOLCHAIN=local`). | Belum |
| R5 | 🟠 Tinggi | `wallet/.../transaction_repository.go` | `GetByWalletID` memfilter kolom `wallet_id` yang tidak ada (seharusnya `sender_wallet_id` / `recipient_wallet_id`). | Belum |
| R6 | 🟠 Tinggi | `auth/internal/usecase/auth_usecase.go` | Pesan error login berbeda (`worng` vs `wrong`) dan tidak ada bcrypt saat email tidak ditemukan → membuka celah *User Enumeration* (via pesan & *timing*). | Belum |
| R7 | 🟠 Tinggi | `auth` handler & usecase | Error database dikirim mentah ke klien dengan status `400`/`401`; seharusnya error internal → `500` dengan pesan generik. Error `Set` Redis diabaikan. | Belum |
| R8 | 🟡 Sedang | `wallet` usecase `TopUp` | `ReferenceID` dibuat acak di server → *Idempotency-Key* belum benar-benar dipakai. | Belum |
| R9 | 🟡 Sedang | `wallet` usecase `GetBalance` | *Lazy init* di endpoint `GET` rawan race (2 request bersamaan → error duplicate). `Currency` & `UpdatedAt` tidak diisi di response. | Belum |
| R10 | 🟡 Sedang | `go.mod` semua service | Versi Go tidak seragam (1.26.0 / 1.25.5 / PRD 1.22+) dan dependensi langsung ditandai `// indirect` → perlu `go mod tidy`. | Belum |
| R11 | 🔵 Desain | `wallet-service` | Validasi sesi JWT via Redis (PRD Bab 2 & 3) belum ada; Redis belum diinisialisasi di `main.go` wallet. | Rencana Modul 3 |
| R12 | 🔵 Desain | `wallet` domain | Ledger masih model 1 baris per transfer; PRD Bab 4 meminta catatan debit/kredit per user + status `PENDING`. | Rencana sebelum Transfer |

---
*(Catatan ini akan diperbarui secara otomatis setiap kali satu langkah selesai).*
