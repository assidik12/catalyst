#!/bin/sh
# Script ini akan menghentikan eksekusi jika terjadi error
set -e

# Menunggu sampai database siap.
# Mekanisme 'healthcheck' di docker-compose adalah cara yang lebih baik,
# namun script ini memberikan kepastian tambahan.
# Pastikan variabel DB_HOST dan DB_PORT tersedia dari .env
# until nc -z $DB_HOST $DB_PORT; do
#   echo "Waiting for database connection..."
#   sleep 2
# done

echo "Running database migrations..."
# Menjalankan perintah migrate.
# -path: lokasi file migrasi di dalam container.
# -database: connection string ke database, diambil dari environment variable.
# 'up': menerapkan semua migrasi yang belum berjalan.
echo "$DB_URL"
migrate -database "$DB_URL" -path db/migrations up

echo "Migrations completed successfully."
echo "Starting the application..."

# Menjalankan aplikasi utama Go yang sudah di-build.
# 'exec' akan menggantikan proses shell dengan aplikasi Go,
# ini adalah best practice agar sinyal (seperti CTRL+C) diteruskan dengan benar.
exec /app/main
