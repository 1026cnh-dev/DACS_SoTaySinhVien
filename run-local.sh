#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"
unset DATABASE_URL
export APP_ENV=local
export CONFIG_FILE=config/local.env
export SQLITE_PATH=student_handbook.db
export UPLOAD_DIR=uploads
export PORT=8080
exec go run .
