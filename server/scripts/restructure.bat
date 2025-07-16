@echo off
cls

echo.
echo ==========================================
echo   MyBlog Project Restructure
echo ==========================================
echo.

REM Check if in correct directory
if not exist "go.mod" (
    echo ERROR: Please run this script from project root directory
    pause
    exit /b 1
)

echo [1/5] Creating backend directory...
mkdir backend 2>nul

echo [2/5] Moving backend files...

REM Move Go related files and directories
if exist "cmd" move cmd backend\ >nul
if exist "internal" move internal backend\ >nul
if exist "pkg" move pkg backend\ >nul
if exist "configs" move configs backend\ >nul
if exist "go.mod" move go.mod backend\ >nul
if exist "go.sum" move go.sum backend\ >nul
if exist "main.go" move main.go backend\ >nul

echo [3/5] Organizing scripts...
mkdir backend\scripts 2>nul
if exist "scripts\watcher.go" move scripts\watcher.go backend\scripts\ >nul

echo [4/5] Creating frontend directory...
mkdir frontend 2>nul

REM Create frontend README
(
echo # Frontend Project
echo.
echo Please put your Vite + SvelteKit project files in this directory.
echo.
echo ## Quick Start
echo.
echo If you don't have a frontend project yet, create one:
echo.
echo ```bash
echo cd frontend
echo npm create svelte@latest .
echo npm install
echo ```
echo.
echo ## Configuration
echo.
echo Make sure to configure the API proxy settings, see `../docs/frontend-integration.md`
) > frontend\README.md

echo [5/5] Creating configuration files...

REM Create .gitignore
(
echo # Go build artifacts
echo *.exe
echo *.exe~
echo *.dll
echo *.so
echo *.dylib
echo /bin/
echo /backend/tmp/
echo.
echo # Test artifacts
echo *.test
echo *.out
echo /coverage/
echo.
echo # Dependencies
echo /vendor/
echo.
echo # Go workspace file
echo go.work
echo.
echo # Environment files
echo .env
echo .env.local
echo .env.development
echo .env.test
echo .env.production
echo.
echo # Log files
echo *.log
echo /logs/
echo.
echo # Database files
echo *.db
echo *.sqlite
echo *.sqlite3
echo.
echo # IDE files
echo .vscode/
echo .idea/
echo *.swp
echo *.swo
echo *~
echo.
echo # OS files
echo .DS_Store
echo .DS_Store?
echo ._*
echo .Spotlight-V100
echo .Trashes
echo ehthumbs.db
echo Thumbs.db
echo.
echo # Temporary files
echo *.tmp
echo *.temp
echo /temp/
echo.
echo # Air hot reload
echo /tmp/
echo air_tmp/
echo.
echo # Config backups
echo *.yaml.bak
echo *.yml.bak
echo *.json.bak
echo.
echo # Build artifacts
echo /dist/
echo /build/
echo.
echo # Certificates
echo *.pem
echo *.key
echo *.crt
echo.
echo # Node.js
echo node_modules/
echo npm-debug.log*
echo yarn-debug.log*
echo yarn-error.log*
echo pnpm-debug.log*
echo lerna-debug.log*
echo.
echo # Frontend build artifacts
echo /frontend/build/
echo /frontend/dist/
echo /frontend/.svelte-kit/
echo.
echo # Frontend environment files
echo /frontend/.env
echo /frontend/.env.local
echo /frontend/.env.development.local
echo /frontend/.env.test.local
echo /frontend/.env.production.local
echo.
echo # Package managers
echo .pnpm-store/
echo .yarn/
echo .npm/
echo.
echo # Docker
echo docker-compose.override.yml
) > .gitignore

REM Create new README
(
echo # MyBlog - Full Stack Blog System
echo.
echo A modern blog system based on Go + SvelteKit.
echo.
echo ## Project Structure
echo.
echo ```
echo MyBlog/
echo ├── backend/          # Go backend service
echo ├── frontend/         # SvelteKit frontend app
echo ├── docs/            # Project documentation
echo ├── scripts/         # Development scripts
echo └── README.md        # Project overview
echo ```
echo.
echo ## Quick Start
echo.
echo ### Development Environment
echo.
echo 1. **Install Dependencies**
echo    - Go 1.20+
echo    - Node.js 18+
echo    - MySQL 8.0+
echo.
echo 2. **Start Services**
echo    ```bash
echo    # Start full development environment
echo    scripts\dev-full.bat
echo    
echo    # Or start separately
echo    cd backend ^&^& go run cmd/myblog/main.go
echo    cd frontend ^&^& npm run dev
echo    ```
echo.
echo 3. **Access Applications**
echo    - Frontend: http://localhost:5173
echo    - Backend API: http://localhost:3000/api
echo.
echo ### Production Deployment
echo.
echo ```bash
echo # Docker way
echo docker-compose up -d
echo.
echo # Manual build
echo scripts\build.bat
echo ```
echo.
echo ## Documentation
echo.
echo - [Frontend Integration Guide](docs/frontend-integration.md)
echo - [API Documentation](docs/api/user_api.md)
echo - [Windows Development](docs/windows-dev.md)
echo.
echo ## Tech Stack
echo.
echo **Backend:**
echo - Go 1.20
echo - Gin Web Framework
echo - GORM + MySQL
echo - Viper Configuration
echo.
echo **Frontend:**
echo - SvelteKit
echo - Vite
echo - TypeScript
echo - TailwindCSS
echo.
echo ## Development Guide
echo.
echo For detailed development guide, see [Frontend Integration Documentation](docs/frontend-integration.md).
) > README.md

echo.
echo ==========================================
echo   Restructure Complete!
echo ==========================================
echo.
echo Summary:
echo   - Backend files moved to backend/ directory
echo   - Created frontend/ directory placeholder
echo   - Updated .gitignore file
echo   - Created new README.md
echo.
echo Next Steps:
echo   1. Copy your frontend project to frontend/ directory
echo   2. Configure vite.config.js proxy settings in frontend
echo   3. Add CORS support to backend
echo   4. Run scripts\dev-full.bat to start full stack environment
echo.
echo For detailed guide, see: docs/frontend-integration.md
echo.
pause