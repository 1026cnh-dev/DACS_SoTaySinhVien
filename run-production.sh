#!/usr/bin/env sh
set -eu
# Dùng để thử cấu hình production trên máy. Secret có thể export từ terminal
# hoặc đặt CONFIG_FILE trỏ tới một file production riêng không commit Git.
APP_ENV=production CONFIG_FILE="${CONFIG_FILE:-config/production.env}" exec go run .
