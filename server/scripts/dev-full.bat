@echo off
cls

echo.
echo ==========================================
echo   MyBlog Full Stack Development
echo ==========================================
echo.

REM Check dependencies
echo [1/6] Checking environment dependencies...

go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed
    echo        Download: https://golang.org/dl/
    pause
    exit /b 1
)

node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Node.js is not installed
    echo        Download: https://nodejs.org/
    pause
    exit /b 1
)

echo      OK: Environment dependencies ready
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
for /f %%i in ('node --version') do set NODE_VERSION=%%i
echo         Go version: %GO_VERSION%
echo         Node version: %NODE_VERSION%
echo.

REM Check project structure
echo [2/6] Checking project structure...

if not exist "backend" (
    echo ERROR: backend directory does not exist
    echo        Please run restructure script first: scripts\restructure.bat
    pause
    exit /b 1
)

if not exist "frontend" (
    echo ERROR: frontend directory does not exist
    echo        Please run restructure script first: scripts\restructure.bat
    pause
    exit /b 1
)

if not exist "frontend\package.json" (
    echo ERROR: Frontend project not initialized
    echo        Please create SvelteKit project in frontend directory
    echo        or copy existing frontend project files to frontend directory
    pause
    exit /b 1
)

echo      OK: Project structure ready
echo.

REM Install backend dependencies
echo [3/6] Installing backend dependencies...
cd backend

if not exist "go.mod" (
    echo ERROR: Backend go.mod file does not exist
    pause
    exit /b 1
)

go mod tidy >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Failed to install backend dependencies
    pause
    exit /b 1
)

echo      OK: Backend dependencies installed
cd ..

REM Install frontend dependencies
echo [4/6] Installing frontend dependencies...
cd frontend

call npm install >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Failed to install frontend dependencies
    pause
    exit /b 1
)

echo      OK: Frontend dependencies installed
cd ..

REM Start backend service
echo [5/6] Starting backend service...
cd backend

echo      Checking database connection...
netstat -an | find "3306" >nul
if %errorlevel% neq 0 (
    echo WARNING: MySQL may not be running on localhost:3306
    echo          Please ensure MySQL service is started
    echo          or check database configuration in backend/configs/config.yaml
)

echo      Starting backend service...
start "MyBlog Backend" /min cmd /c "go run cmd/myblog/main.go > ..\backend.log 2>&1"

REM Wait for backend to start
timeout /t 3 /nobreak >nul

echo      Checking backend service status...
set BACKEND_READY=0
for /l %%i in (1,1,10) do (
    curl -s http://localhost:3000/api/health >nul 2>&1
    if !errorlevel! equ 0 (
        set BACKEND_READY=1
        goto backend_ready
    )
    timeout /t 1 /nobreak >nul
)

:backend_ready
if %BACKEND_READY% equ 1 (
    echo      OK: Backend service started successfully
    echo         Address: http://localhost:3000
) else (
    echo ERROR: Backend service failed to start
    echo        Please check backend.log file
    pause
    exit /b 1
)

cd ..

REM Start frontend service
echo [6/6] Starting frontend service...
cd frontend

echo      Starting frontend service...
start "MyBlog Frontend" /min cmd /c "npm run dev > ..\frontend.log 2>&1"

REM Wait for frontend to start
timeout /t 3 /nobreak >nul

echo      Checking frontend service status...
set FRONTEND_READY=0
for /l %%i in (1,1,10) do (
    curl -s http://localhost:5173 >nul 2>&1
    if !errorlevel! equ 0 (
        set FRONTEND_READY=1
        goto frontend_ready
    )
    timeout /t 1 /nobreak >nul
)

:frontend_ready
if %FRONTEND_READY% equ 1 (
    echo      OK: Frontend service started successfully
    echo         Address: http://localhost:5173
) else (
    echo WARNING: Frontend service may still be starting
    echo          Please wait a moment or check frontend.log file
)

cd ..

echo.
echo ==========================================
echo   Full Stack Environment Started!
echo ==========================================
echo.
echo Service Addresses:
echo   Frontend App: http://localhost:5173
echo   Backend API:  http://localhost:3000/api
echo   Health Check: http://localhost:3000/api/health
echo.
echo Log Files:
echo   Backend log: backend.log
echo   Frontend log: frontend.log
echo.
echo To Stop Services:
echo   Run: scripts\stop.bat
echo   Or close the backend/frontend command windows
echo.
echo Tips:
echo   - Backend code changes require manual restart
echo   - Frontend code changes will auto-reload
echo   - Check log files if services don't work properly
echo.

REM Create stop script
(
echo @echo off
echo echo Stopping MyBlog development services...
echo echo.
echo taskkill /f /im "cmd.exe" /fi "WINDOWTITLE eq MyBlog Backend*" ^>nul 2^>^&1
echo if %%errorlevel%% equ 0 ^(
echo     echo Backend service stopped
echo ^) else ^(
echo     echo Backend service was not running
echo ^)
echo.
echo taskkill /f /im "cmd.exe" /fi "WINDOWTITLE eq MyBlog Frontend*" ^>nul 2^>^&1
echo if %%errorlevel%% equ 0 ^(
echo     echo Frontend service stopped
echo ^) else ^(
echo     echo Frontend service was not running
echo ^)
echo.
echo if exist backend.log del backend.log
echo if exist frontend.log del frontend.log
echo echo Log files cleaned
echo.
echo echo All services stopped
echo pause
) > scripts\stop.bat

echo Stop script created: scripts\stop.bat
echo.
echo Happy Coding!
echo.
pause