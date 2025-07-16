@echo off
cls

echo.
echo ==========================================
echo   MyBlog Development Environment
echo ==========================================
echo.

REM Check Go installation
echo [1/4] Checking Go environment...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo.
    echo ERROR: Go is not installed
    echo        Download: https://golang.org/dl/
    echo.
    pause
    exit /b 1
)
echo      OK: Go environment ready
echo.

REM Clean temporary files
echo [2/4] Cleaning temporary files...
if exist tmp rmdir /s /q tmp >nul 2>&1
mkdir tmp >nul 2>&1
echo      OK: Temporary files cleaned
echo.

REM Check dependencies
echo [3/4] Checking dependencies...
go mod tidy >nul 2>&1
if %errorlevel% neq 0 (
    echo.
    echo ERROR: Failed to install dependencies
    echo.
    pause
    exit /b 1
)
echo      OK: Dependencies ready
echo.

echo [4/4] Starting hot reload service...
echo.
echo ==========================================
echo   Hot Reload Started
echo ==========================================
echo.
echo Features:
echo   - Auto watch .go files
echo   - Auto compile on changes  
echo   - Server: http://localhost:3000
echo.
echo Press Ctrl+C to stop
echo.
echo ------------------------------------------

REM Start file watcher
go run scripts/watcher.go