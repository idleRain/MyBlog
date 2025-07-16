@echo off
cls

echo MyBlog Hot Reload Starting...
echo.

go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go not installed
    echo Download: https://golang.org/dl/
    pause
    exit /b 1
)

echo Go environment OK
echo.

if exist tmp rmdir /s /q tmp >nul 2>&1
mkdir tmp >nul 2>&1

echo Checking dependencies...
go mod tidy >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Failed to install dependencies
    pause
    exit /b 1
)

echo Dependencies OK
echo.
echo Starting hot reload...
echo Server: http://localhost:3000
echo Press Ctrl+C to stop
echo.

go run scripts/watcher.go