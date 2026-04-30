# Point of Sale (POS) API

Aplikasi ini adalah backend RESTful API untuk Sistem Point of Sale (POS) atau Kasir. Dibangun menggunakan arsitektur modern berbasis Go (Golang) dan Fiber framework, aplikasi ini dirancang untuk performa tinggi, skalabilitas, dan kemudahan pemeliharaan.

## 📋 Daftar Isi
- [Teknologi & Library](#-teknologi--library)
- [Arsitektur & Struktur Folder](#-arsitektur--struktur-folder)
- [Skema Database](#-skema-database)
- [Daftar API (Endpoints)](#-daftar-api-endpoints)
- [Persiapan & Instalasi (Setup)](#-persiapan--instalasi-setup)
- [Cara Menjalankan Aplikasi](#-cara-menjalankan-aplikasi)
- [Cara Menjalankan Testing](#-cara-menjalankan-testing)

---

## 🚀 Teknologi & Library

**Tech Stack Utama:**
- **Bahasa Pemrograman:** [Go (Golang)](https://go.dev/) v1.25.0
- **Web Framework:** [Fiber v2](https://gofiber.io/) (Cepat dan efisien)
- **Database:** PostgreSQL
- **ORM:** [GORM](https://gorm.io/)
- **Database Migration:** [Atlas](https://atlasgo.io/) & [golang-migrate](https://github.com/golang-migrate/migrate)

**Library Tambahan Terpenting:**
- `golang-jwt/jwt/v4` - Untuk Autentikasi berbasis JSON Web Token (JWT).
- `crypto/bcrypt` - Untuk hashing password dengan aman.
- `swaggo/swag` & `arsmn/fiber-swagger` - Untuk dokumentasi API terotomatisasi (Swagger UI).
- `go-playground/validator/v10` - Untuk validasi payload request.
- `boombuler/barcode` - Untuk pembuatan barcode produk.
- `testify` - Untuk mempermudah proses unit testing.

---

## 📁 Arsitektur & Struktur Folder

Aplikasi ini mengadopsi struktur standar yang rapi (mirip *Clean Architecture* sederhana) untuk memisahkan antara _routing_, logika HTTP, logika bisnis, dan interaksi ke database.

```text
pos-api/
├── cmd/                # Entry point utama aplikasi
│   ├── api/            # Entry point untuk menjalankan server HTTP utama
│   └── seeder/         # Script untuk memasukkan dummy data (seeding)
├── configs/            # Pengaturan konfigurasi (misal: parsing .env)
├── database/           # Setup koneksi database & file migrasi Atlas/golang-migrate
├── docs/               # Berisi file hasil generate Swagger API documentation
├── internal/           # Kode inti aplikasi (tidak bisa di-import dari project lain)
│   ├── handlers/       # HTTP Controller: Menerima request, memanggil service, mengirim response
│   ├── middlewares/    # Interceptor request (contoh: JWT Auth, RBAC Admin/Cashier)
│   ├── models/         # Definisi struct/entity database (GORM schemas)
│   ├── repositories/   # Lapisan akses database (CRUD) menggunakan GORM
│   ├── routes/         # Pemetaan URL path ke Handler (Public & Protected routes)
│   └── services/       # Lapisan logika bisnis (Business logic layer)
├── pkg/                # Library/Utility internal yang bisa digunakan ulang (helpers)
├── tests/              # Berisi integration test atau unit test
├── .env                # File konfigurasi environment (kredensial DB, Secret JWT, dll)
├── atlas.hcl           # Konfigurasi deklaratif untuk migrasi database dengan Atlas
├── go.mod & go.sum     # Dependency manager Go
└── MIGRATION.md        # Panduan teknis cara melakukan migrasi database
```

---

## 🗄️ Skema Database

Sistem ini memiliki beberapa entitas (tabel) utama yang saling berelasi:

1. **`users`**: Menyimpan data pengguna aplikasi beserta _Role_ mereka (`admin`, `manager`, `kasir`). Password disimpan dalam bentuk hash (bcrypt).
2. **`categories`**: Kategori pengelompokan produk.
3. **`products`**: Menyimpan data master barang, termasuk harga, SKU/Barcode, dan jumlah stok saat ini. Berelasi dengan tabel `categories`. Mendukung *soft-delete*.
4. **`inventory_logs`**: Mencatat histori pergerakan stok barang. Setiap penambahan atau pengurangan produk (baik manual maupun via transaksi) akan tercatat di sini.
5. **`transactions`**: Header dari sebuah transaksi penjualan. Menyimpan kasir yang bertugas, metode pembayaran, total bayar, tanggal, dan status (Selesai, Batal, Retur).
6. **`transaction_details`**: Item yang dibeli dalam sebuah transaksi. Berelasi dengan `transactions` dan `products`. Menyimpan harga saat pembelian (agar jika harga produk berubah, histori transaksi tetap aman).
7. **`payment_methods`**: Metode pembayaran yang didukung toko (Cash, QRIS, Transfer, dsb).
8. **`cash_flows`**: Buku kas toko. Mencatat Pemasukan (Income), Pengeluaran (Outcome), dan Modal Awal (Capital). Terhubung dengan transaksi (penjualan menambah income).
9. **`store_settings`**: Menyimpan konfigurasi global toko (Nama Toko, Alamat, Teks Struk/Footer).

---

## 🌐 Daftar API (Endpoints)

Sistem menggunakan standar RESTful. Sebagian besar API di bawah ini diamankan dengan JWT dan dibatasi berdasarkan Role (RBAC).
Untuk dokumentasi lengkap beserta request/response bodynya, silakan akses **Swagger UI** saat aplikasi berjalan: `http://localhost:8000/swagger/`

**Ringkasan Endpoint:**

*   **Autentikasi:**
    *   `POST /api/v1/auth/login` - Mendapatkan JWT token.
    *   `GET /api/v1/auth/profile` - Mengambil data user yang sedang login.
*   **Dashboard & Reports (Admin/Manager):**
    *   `GET /api/v1/dashboard/` - Statistik ringkas toko.
    *   `GET /api/v1/reports/sales` - Laporan penjualan terperinci.
*   **Products & Categories:**
    *   `GET, POST, PUT, DELETE /api/v1/products` - CRUD produk.
    *   `GET /api/v1/products/low-stock` - Mengambil produk yang perlu di-restock.
    *   `GET, POST, PUT, DELETE /api/v1/categories` - CRUD kategori produk.
*   **Transactions (POS):**
    *   `POST /api/v1/transactions` - Membuat transaksi baru (Checkout kasir).
    *   `GET /api/v1/transactions` - Riwayat transaksi.
    *   `POST /api/v1/transactions/:id/cancel` - Membatalkan transaksi.
*   **Inventory:**
    *   `GET /api/v1/inventory` - Log pergerakan inventori.
    *   `POST /api/v1/inventory` - Penyesuaian stok (Adjust stock) manual.
*   **Cash Flow:**
    *   `GET, POST, PUT, DELETE /api/v1/cash-flow` - Mengatur buku kas.
*   **Store Settings & Payment Methods:**
    *   `GET, PUT /api/v1/store-settings` - Pengaturan toko.
    *   `GET, POST, PUT, DELETE /api/v1/payment-methods` - Mengelola tipe pembayaran.
*   **Barcode & Export:**
    *   `GET /api/v1/barcode/:id` - Generate barcode gambar.
    *   `GET /api/v1/export/products/csv` - Export data ke CSV.

---

## 🛠️ Persiapan & Instalasi (Setup)

1. **Clone repositori** atau pastikan Anda berada di root direktori project `pos-api`.
2. **Copy file Environment**:
   Pastikan Anda menyalin atau membuat file `.env` dari template yang disediakan (contoh: `.env.example`).
   ```bash
   cp .env.example .env
   ```
3. **Isi konfigurasi `.env`**:
   Atur koneksi database PostgreSQL Anda sesuai dengan kredensial yang diinginkan. File `.env` ini akan dibaca secara otomatis oleh Docker maupun aplikasi lokal.
   ```env
   DB_HOST=localhost # Gunakan 'db' jika menggunakan Docker
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=password_anda
   DB_NAME=pos_db
   JWT_SECRET=super_rahasia_sekali
   PORT=8000
   ```
4. **Unduh Dependencies (Library)**:
   ```bash
   go mod download
   go mod tidy
   ```
5. **Siapkan Database PostgreSQL**:
   Pastikan PostgreSQL server sudah berjalan dan database kosong dengan nama `pos_db` (atau sesuai `.env`) sudah dibuat.
6. **Jalankan Migrasi Database**:
   Aplikasi ini menggunakan skema migrasi berbasis Atlas/golang-migrate. Silakan merujuk pada file `MIGRATION.md` untuk panduan langkah demi langkah memigrasikan skema ke database Anda.

---

## ▶️ Cara Menjalankan Aplikasi

Anda dapat menjalankan aplikasi ini menggunakan **Docker** (Sangat Disarankan) atau secara lokal.

### Opsi A: Menjalankan dengan Docker Compose
Cara ini sangat mudah karena secara otomatis menjalankan server API beserta database PostgreSQL.
1. Pastikan Anda telah mensetup `.env`.
2. Jalankan perintah:
   ```bash
   docker-compose up -d --build
   ```
3. API akan berjalan di port `8080` (sesuai konfigurasi docker-compose). Untuk melihat log: `docker-compose logs -f app`.

### Opsi B: Menjalankan secara Lokal
1. Pastikan PostgreSQL sudah jalan di komputer Anda.
2. Jalankan perintah:
   ```bash
   go run cmd/api/main.go
   ```

### 🗃️ Memasukkan Data Awal (Seeding) - Opsional
Jika ini pertama kalinya Anda menjalankan aplikasi dan butuh data awalan (seperti akun Admin/Kasir default, kategori dummy, atau produk dummy), jalankan seeder berikut.

*Jika via lokal:*
```bash
go run cmd/seeder/main.go
```
*Jika via Docker:*
```bash
docker-compose exec app go run cmd/seeder/main.go
```

---

## 🧪 Cara Menjalankan Testing

Proyek ini telah dikonfigurasi untuk menjalankan unit dan integration testing bawaan Golang.

Untuk menjalankan semua test di dalam aplikasi:
```bash
go test ./... -v
```

Untuk melihat cakupan kode tes (Code Coverage):
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```
Tindakan ini akan membuka browser dan menunjukkan baris kode mana saja yang sudah atau belum ter-cover oleh unit test.
