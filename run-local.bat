@echo off
setlocal
cd /d "%~dp0"
set APP_ENV=local
set CONFIG_FILE=config\local.env
set DATABASE_URL=
set SQLITE_PATH=student_handbook.db
set UPLOAD_DIR=uploads
set PORT=8080
go run .
endlocal
