# Panduan Install dan Setup Service

Dokumen ini menjelaskan cara install, konfigurasi, dan menjalankan aplikasi backup PostgreSQL ke S3/MinIO sebagai service.

## Ringkasan Cara Kerja

Aplikasi akan:

1. Membaca konfigurasi dari file `.env`
2. Menjalankan `pg_dump`
3. Upload hasil dump ke object storage (S3/MinIO)
4. Membersihkan (menghapus) file backup lama yang sudah berusia lebih dari 10 hari
5. Mengulang proses tiap 12 jam

## 1) Prasyarat

- Go versi 1.24 atau lebih baru
- PostgreSQL client tools (wajib ada `pg_dump`)
- Akses ke database PostgreSQL
- Akses ke bucket S3/MinIO

Catatan penting:

- Kode saat ini menggunakan koneksi TLS ke object storage (`useSSL = true`), jadi endpoint harus mendukung HTTPS.
- Variabel password database bisa pakai `DB_PASS` atau `DB_PASSWORD`.

## 2) Install Dependency Aplikasi

Jalankan dari root project:

```bash
go mod tidy
go build -o backup .
```

Untuk Windows:

```powershell
go mod tidy
go build -o backup.exe .
```

## 3) Konfigurasi Environment

1. Salin file contoh:

```bash
cp .env.example .env
```

Jika di PowerShell:

```powershell
Copy-Item .env.example .env
```

2. Isi nilai pada `.env`:

```env
DB_HOST=localhost
DB_USER=postgres
DB_PASS=your_db_password
DB_NAME=your_database
DB_PORT=5432

S3_URL=s3.your-domain.com
S3_BUCKET=your-backup-bucket
S3_ACCESS_KEY_ID=your-access-key
S3_SECRET_ACCESS_KEY=your-secret-key
S3_KEY_PREFIX=backup/postgres

# Windows contoh:
PG_DUMP_PATH="C:/Program Files/PostgreSQL/17/bin/pg_dump.exe"

# Linux contoh:
# PG_DUMP_PATH=/usr/bin/pg_dump
```

Keterangan variabel:

- `S3_URL`: endpoint tanpa skema, contoh `s3.your-domain.com` atau `minio.your-domain.com`.
- `S3_KEY_PREFIX`: folder prefix object di bucket (opsional).
- `PG_DUMP_PATH`: path binary `pg_dump`. Jika kosong, aplikasi mencoba `pg_dump` dari PATH.

## 4) Uji Jalankan Manual

Sebelum dijadikan service, jalankan manual dulu:

```bash
./backup
```

Windows:

```powershell
.\backup.exe
```

Jika sukses, log akan menampilkan proses backup lalu upload berhasil.

## 5) Setup Service Linux (systemd)

### 5.1 Siapkan direktori deploy

Contoh:

- binary: `/home/nama/app/backup-database-using-golang/backup`
- env: `/home/nama/app/backup-database-using-golang/.env`

Perintah contoh:

```bash
sudo mkdir -p /home/nama/app/backup-database-using-golang
sudo cp backup /home/nama/app/backup-database-using-golang/backup
sudo cp .env /home/nama/app/backup-database-using-golang/.env
sudo chmod 600 /home/nama/app/backup-database-using-golang/.env
sudo chmod +x /home/nama/app/backup-database-using-golang/backup
```

### 5.2 Buat user service (opsional, direkomendasikan)

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin nama || true
sudo chown -R nama:nama /home/nama/app/backup-database-using-golang
```

### 5.3 Buat unit file systemd

Buat file `/etc/systemd/system/db-backup.service`:

```ini
[Unit]
Description=Automatic Database Backup to MinIO
After=network.target postgresql.service

[Service]
User=nama
WorkingDirectory=/home/nama/app/backup-database-using-golang
ExecStart=/home/nama/app/backup-database-using-golang/backup
EnvironmentFile=/home/nama/app/backup-database-using-golang/.env
Restart=always
RestartSec=10
StandardOutput=append:/var/log/db-backup.log
StandardError=append:/var/log/db-backup-error.log
Environment="PATH=/usr/local/bin:/usr/bin:/bin:/usr/local/go/bin"

[Install]
WantedBy=multi-user.target
```

### 5.4 Aktifkan service

```bash
sudo systemctl daemon-reload
sudo systemctl enable db-backup
sudo systemctl start db-backup
sudo systemctl status db-backup
```

Lihat log:

```bash
journalctl -u db-backup -f
```

## 6) Setup Service Windows (opsi cepat dengan NSSM)

Jika ingin jalan sebagai Windows Service, cara praktis adalah pakai NSSM.

1. Download NSSM
2. Install service:

```powershell
nssm install SqlBackupService "D:\path\to\backup.exe"
```

3. Set:

- Startup directory: folder aplikasi (berisi `.env`)
- AppEnvironmentExtra: bisa isi variabel env jika tidak pakai file `.env`

4. Start service:

```powershell
nssm start SqlBackupService
```

## 7) Troubleshooting

- Error `DB_NAME is required`:
  - Isi `DB_NAME` di `.env`.
- Error `failed to upload to object storage`:
  - Cek kredensial S3, bucket, dan endpoint.
  - Pastikan endpoint mendukung HTTPS.
- Error `pg_dump failed`:
  - Cek `PG_DUMP_PATH`.
  - Pastikan user database punya akses dump.
- Service start tapi backup gagal:
  - Cek log service (`journalctl` di Linux atau Event Viewer/NSSM logs di Windows).

## 8) Rekomendasi Operasional

- Simpan `.env` dengan permission ketat (jangan expose secret).
- Uji restore database secara berkala untuk validasi backup.
