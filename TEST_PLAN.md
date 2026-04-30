# API Unit Testing Implementation Plan

**Objective:**
Dokumen ini merupakan panduan/perencanaan (*blueprint*) bagi model AI selanjutnya untuk mengimplementasikan *Unit Test* dan *Integration Test* untuk semua endpoint API di backend sistem POS. 

---

## 🏗️ Aturan Utama (Ground Rules)
1. **Lokasi File Test:** Semua file test harus dibuat di dalam folder `tests/` (misalnya `tests/integration/` atau `tests/api/`).
2. **Library Testing:** Gunakan bawaan Go `testing` dipadukan dengan `github.com/stretchr/testify` (untuk `assert` dan `require`) atau *mocking library* yang sudah ada.
3. **Isolasi Data (Sangat Penting):** 
   Setiap *test case* atau blok skenario **WAJIB menghapus data dari tabel terkait (Database Cleanup)** sebelum eksekusi dimulai. Hal ini untuk memastikan tidak ada data residu yang membuat hasil pengujian menjadi tidak konsisten (*flaky tests*).
4. **Detail Implementasi:** Anda (Model AI pengimplementasi) diberikan kebebasan penuh untuk menulis logika inisialisasi server (Fiber app), *mock* database (atau DB *in-memory* SQLite/Testcontainers), serta *dummy payload*. Fokuslah pada *coverage* skenario di bawah ini.

---

## 🎯 Daftar Skenario Pengujian (Test Scenarios)

### 1. Modul Autentikasi (Auth)
*   **`POST /api/v1/auth/login`**
    *   Berhasil login dengan kredensial yang benar (Mengembalikan JWT).
    *   Gagal login karena *password* salah.
    *   Gagal login karena *username* tidak ditemukan.
*   **`GET /api/v1/auth/profile`**
    *   Berhasil mendapatkan profil ketika membawa JWT valid.
    *   Ditolak (401) ketika token tidak ada atau invalid.
*   **`PUT /api/v1/auth/profile` & `PUT /api/v1/auth/password`**
    *   Berhasil mengubah profil dan password.
    *   Gagal karena payload tidak valid.

### 2. Modul Produk & Kategori (Products & Categories)
*   **`POST /api/v1/categories` & `POST /api/v1/products`**
    *   Admin berhasil membuat produk/kategori baru.
    *   Kasir ditolak (403 Forbidden) saat mencoba membuat produk.
    *   Gagal karena validasi *mandatory fields* (misal: nama atau harga kosong).
*   **`GET /api/v1/products` & `GET /api/v1/categories`**
    *   Berhasil mengambil daftar produk beserta filternya (pagination berjalan benar).
*   **`PUT /api/v1/products/:id` & `DELETE /api/v1/products/:id`**
    *   Berhasil memperbarui data produk.
    *   Berhasil melakukan *soft-delete* produk.
    *   Berhasil melakukan *restore* produk yang ter-*soft-delete*.
*   **`GET /api/v1/products/low-stock`**
    *   Berhasil menampilkan list produk yang stoknya di bawah batas peringatan.

### 3. Modul Transaksi POS (Transactions)
*   **`POST /api/v1/transactions` (Checkout)**
    *   Transaksi berhasil: Stok produk terpotong, data masuk ke `transactions`, `transaction_details`, dan menghasilkan *domain event* Arus Kas (Pemasukan).
    *   Gagal: Karena stok barang tidak mencukupi (harus *rollback* DB transaction).
    *   Gagal: Memilih metode pembayaran yang tidak valid/tidak aktif.
*   **`GET /api/v1/transactions` & `GET /api/v1/transactions/:id`**
    *   Admin/Manager berhasil melihat riwayat dan detail transaksi.
*   **`POST /api/v1/transactions/:id/cancel` & `Return`**
    *   Admin berhasil membatalkan transaksi: Pastikan status transaksi berubah, stok barang kembali ke inventori, dan arus kas tercatat keluar (Refund).

### 4. Modul Inventori (Inventory)
*   **`POST /api/v1/inventory` (Adjust Stock)**
    *   Berhasil menambah stok produk secara manual (Log inventori tercatat).
    *   Berhasil mengurangi stok produk karena rusak/hilang.
*   **`GET /api/v1/inventory` & `GET /api/v1/inventory/stats`**
    *   Berhasil mengambil log riwayat pergerakan barang secara berurutan.

### 5. Modul Arus Kas (Cash Flow)
*   **`POST /api/v1/cash-flow`**
    *   Berhasil memasukkan Modal Awal (Capital).
    *   Berhasil mencatat Pengeluaran manual (Outcome).
*   **`GET /api/v1/cash-flow/summary`**
    *   Verifikasi bahwa perhitungan Pemasukan - Pengeluaran + Modal menghasilkan saldo akhir (Balance) yang akurat.

### 6. Modul Pengaturan & Metode Pembayaran (Settings & Payments)
*   **`GET, PUT /api/v1/store-settings`**
    *   Membaca dan memperbarui identitas toko berhasil.
*   **`GET, POST, PUT, DELETE /api/v1/payment-methods`**
    *   Menambah metode pembayaran baru.
    *   Menguji endpoint `GET /active` hanya mengembalikan payment method yang berstatus aktif.

### 7. Laporan & Export (Reports & Export)
*   **`GET /api/v1/reports/...`**
    *   Memastikan format laporan JSON kembaliannya tepat (Sales, Products, Stock-Value).
*   **`GET /api/v1/export/...`**
    *   Memastikan endpoint mengembalikan header Content-Type CSV (`text/csv`) dan datanya diformat dengan pemisah koma.

---

**Catatan untuk AI Pengimplementasi:**
Fokus pada kejelasan dan kemudahan *maintenance* dari kode test. Buat fungsi bantuan (*helper functions*) untuk proses seperti login dan mengambil token, atau membersihkan database (misal `TruncateTables(db)`) sebelum eksekusi blok skenario. Eksekusilah rencana ini file-per-file jika terlalu besar.
