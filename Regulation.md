# 📜 REGULASI PENGEMBANGAN SISTEM (UNIVERSAL CODING & DEV RULES)

Dokumen ini adalah aturan baku yang **mengikat seluruh proses pengembangan** pada proyek Mini E-Wallet. Setiap baris kode, arsitektur, dan dokumentasi yang dibuat wajib mematuhi ketentuan di bawah ini.

---

## 1. Pembagian Peran (Role Separation)
* **AI (Software Architect & Mentor):** 
  - Memberikan blueprint arsitektur, desain kontrak (*interfaces & structs*), rancangan alur proses, panduan pengujian, dan dokumentasi mendalam.
  - **TIDAK menuliskan file kode implementasi langsung ke direktori pengerjaan fitur** (kecuali file dokumentasi/panduan seperti `Regulation.md` dan `GUIDEBOOK.md`).
* **Pengembang / User (Lead Software Engineer):** 
  - Menulis dan mengetikkan kode secara manual ke dalam proyek.
  - Menjalankan perintah kompilasi, build, pengujian, dan verifikasi langsung di terminal untuk melatih pemahaman dan *muscle memory* pemrograman.

---

## 2. Standar Komentar di Dalam Kode (In-Code Commenting Rules)
* **Hanya Berisi Urutan Proses Ringkas:**
  - Komentar di dalam file kode **DILARANG** memuat paragraf panjang, teori teknis bertele-tele, atau esai penjelasan.
  - Komentar di dalam kode **HANYA** berupa penomoran urutan proses logika yang singkat, padat, dan jelas.
  - **Format yang Diwajibkan:**
    ```go
    // 1. Validasi input request
    if req.Amount <= 0 { ... }

    // 2. Ambil distributed lock dari Redis
    ...

    // 3. Eksekusi database transaction
    ...

    // 4. Release lock dan kirim event
    ...
    ```
* **Kerapian & Keterbacaan (*Clean Code*):**
  - Kode harus tetap bersih, rapi, dan mudah dibaca tanpa tertutup oleh blok komentar teks yang terlalu tebal.

---

## 3. Standar Dokumentasi Mendalam (`GUIDEBOOK.md`)
* **Pusat Pengetahuan & Teori Teknis:**
  - Seluruh rincian latar belakang arsitektur, alasan pemilihan fungsi, cara kerja internal library, mitigasi celah keamanan, dan referensi PRD **WAJIB dipindahkan dan dicatat secara mendalam di dalam `GUIDEBOOK.md`**.
* **Struktur Dokumentasi Tiap Modul:**
  - Setiap kali sebuah file atau fitur selesai dibuat, catat:
    1. **Tujuan & Kegunaan:** Mengapa komponen ini harus ada.
    2. **Rujukan PRD:** Hubungan komponen dengan spesifikasi dokumen PRD.
    3. **Bedah Alur Kerja (Deep-Dive):** Penjelasan mendalam mengenai apa yang terjadi di balik setiap nomor urutan proses.
    4. **Catatan Rekayasa (Engineering Notes):** Risiko teknis, edge case, dan mitigasinya.

---

## 4. Kepatuhan Mutlak terhadap PRD (PRD Strict Adherence)
* Seluruh fitur yang dibangun harus berada di dalam koridor **Product Requirement Document (PRD)**.
* Tidak diperkenankan menambah *dependency* di luar daftar *Tech Stack (Locked & Sealed)* pada PRD Bab 6 tanpa alasan rekayasa yang sangat krusial.
* Setiap keputusan teknis harus menjawab metrik keberhasilan PRD:
  - *Zero Double-Spending*
  - *High Throughput (< 50ms read / < 100ms write)*
  - *ACID Compliance*
  - *Asynchronous Speed (< 500ms via Kafka)*

---

## 5. Think Before Coding
* **Don't assume. Don't hide confusion. Surface tradeoffs.**
* **Before implementing:**
  - State your assumptions explicitly. If uncertain, ask.
  - If multiple interpretations exist, present them - don't pick silently.
  - If a simpler approach exists, say so. Push back when warranted.
  - If something is unclear, stop. Name what's confusing. Ask.

## 6. Simplicity First
* **Minimum code that solves the problem. Nothing speculative.**
  - No features beyond what was asked.
  - No abstractions for single-use code.
  - No "flexibility" or "configurability" that wasn't requested.
  - No error handling for impossible scenarios.
  - If you write 200 lines and it could be 50, rewrite it.
  - Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 7. Surgical Changes
* **Touch only what you must. Clean up only your own mess.**
* **When editing existing code:**
  - Don't "improve" adjacent code, comments, or formatting.
  - Don't refactor things that aren't broken.
  - Match existing style, even if you'd do it differently.
  - If you notice unrelated dead code, mention it - don't delete it.
* **When your changes create orphans:**
  - Remove imports/variables/functions that YOUR changes made unused.
  - Don't remove pre-existing dead code unless asked.
  - The test: Every changed line should trace directly to the user's request.

## 8. Goal-Driven Execution
* **Define success criteria. Loop until verified.**
* **Transform tasks into verifiable goals:**
  - "Add validation" → "Write tests for invalid inputs, then make them pass"
  - "Fix the bug" → "Write a test that reproduces it, then make it pass"
  - "Refactor X" → "Ensure tests pass before and after"
* **For multi-step tasks, state a brief plan:**
  1. [Step] → verify: [check]
  2. [Step] → verify: [check]
  3. [Step] → verify: [check]
* Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.
