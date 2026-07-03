# WA AI Assistant

WhatsApp AI Assistant & Blast — auto-reply pesan WhatsApp dengan AI (kompatibel OpenAI,
mis. DeepSeek) berbasis knowledge base, plus blast/broadcast dengan pengaman anti-blokir.

## Auto-reply AI
- Balas pesan WhatsApp otomatis dengan AI.
- **Knowledge base**: isi tanya-jawab manual, generate otomatis dengan AI, atau impor.
- **Web crawler**: latih AI dari isi website (crawl → pilih halaman → train).
- **Semantic search** (opsional) via embedding untuk jawaban lebih relevan.
- Pilihan **tone**: ramah, formal, santai, persuasif.
- **Handoff**: alihkan percakapan ke manusia bila perlu.

## Blast / Broadcast
- Kirim pesan massal ke daftar nomor dengan **teks + lampiran** (gambar/video/dokumen).
- **Jadwal blast** ke tanggal & jam tertentu.
- **Blast ke grup**: post satu pesan ke banyak grup sekaligus (terjadwal).
- **Personalisasi** `{nama}` per penerima.
- Ambil penerima dari: pernah chat, kontak WA, anggota grup, atau label.
- **Cek nomor terdaftar WhatsApp** sebelum kirim, untuk membuang nomor tidak aktif.

### Pengaman anti-blokir
- **Jeda acak** antar pesan + **istirahat berkala** agar pola kirim tidak seperti bot.
- **Humanized typing** (indikator "mengetik…").
- **Opt-out otomatis**: kontak yang balas STOP/BERHENTI dilewati.
- **Consent tracking** per kategori pesan + **risk level** sebelum blast.
- Lanjut otomatis bila server restart; jeda otomatis bila WhatsApp membatasi.

## Manajemen grup (Anti-Spam)
- Moderasi grup: deteksi link/nomor/kata terlarang & flood.
- Aksi: hapus pesan, tandai untuk dikeluarkan, atau auto-kick (butuh bot admin).
- Log audit tiap tindakan moderasi.

## CRM & kontak
- Simpan kontak, beri tag/label, impor massal.
- Riwayat chat & analitik percakapan.
- Follow-up otomatis bertahap (multi-step).
- Formulir closing & pencatatan order; cek ongkir (opsional).

## Integrasi & tracking
- **Meta Conversions API (CAPI)**: kirim event konversi ke Meta (rahasia dienkripsi at-rest).
- Google Sheets (opsional).

## Multi-agent
- Kelola beberapa nomor/agent WhatsApp dalam satu dashboard.

## Keamanan & operasional
- Login admin dengan password ter-hash (bcrypt) + throttle/lockout.
- JWT untuk sesi; rahasia sensitif dienkripsi di database.
- Sistem lisensi (aktivasi + heartbeat + grace offline).

---
Penggunaan tunduk pada [EULA](docs/EULA.md) & [Disclaimer](docs/DISCLAIMER.md). **Wajib dibaca.**
