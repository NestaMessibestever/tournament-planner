# Tournament Planner Complete Setup Script for Windows
# This script sets up the entire project from scratch
# Run as Administrator in PowerShell

param(
    [string]$ProjectPath = "D:\TournamentPlanner"
)

# Color functions for better output
function Write-Step { param($message) Write-Host "`n=== $message ===" -ForegroundColor Green }
function Write-Info { param($message) Write-Host "INFO: $message" -ForegroundColor Cyan }
function Write-Success { param($message) Write-Host "[OK] $message" -ForegroundColor Green }
function Write-Error { param($message) Write-Host "ERROR: $message" -ForegroundColor Red }
function Write-Warning { param($message) Write-Host "WARNING: $message" -ForegroundColor Yellow }

# Check if running as Administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Error "This script must be run as Administrator. Exiting..."
    exit 1
}

Write-Host @"
╔═══════════════════════════════════════════════════════════════╗
║          Tournament Planner - Native Windows Setup            ║
║                                                               ║
║  This script will set up the complete Tournament Planner      ║
║  application on your Windows 11 machine without Docker.       ║
║                                                               ║
║  Prerequisites that should already be installed:              ║
║  - Go 1.21+                                                  ║
║  - Node.js 18+                                               ║
║  - MySQL 8.0                                                 ║
║  - MongoDB 6.0                                               ║
║  - Redis/Memurai                                             ║
╚═══════════════════════════════════════════════════════════════╝
"@ -ForegroundColor Cyan

# Confirm prerequisites
$confirm = Read-Host "`nHave you installed all prerequisites? (yes/no)"
if ($confirm -ne "yes") {
    Write-Warning "Please install all prerequisites first, then run this script again."
    exit 0
}

# Step 1: Create Project Structure
Write-Step "Creating Project Structure at $ProjectPath"

try {
    # Create main directory
    New-Item -ItemType Directory -Force -Path $ProjectPath | Out-Null
    Set-Location $ProjectPath
    
    # Create all backend directories
    $backendDirs = @(
        "backend\cmd\server",
        "backend\internal\api",
        "backend\internal\config", 
        "backend\internal\database",
        "backend\internal\middleware",
        "backend\internal\models",
        "backend\internal\repositories",
        "backend\internal\services",
        "backend\internal\server",
        "backend\internal\websocket",
        "backend\internal\utils",
        "backend\migrations",
        "backend\uploads",
        "backend\tmp",
        "backend\pkg"
    )
    
    foreach ($dir in $backendDirs) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }
    
    # Create all frontend directories
    $frontendDirs = @(
        "frontend\public",
        "frontend\src\components\layout",
        "frontend\src\components\ui",
        "frontend\src\components\tournaments",
        "frontend\src\components\common",
        "frontend\src\contexts",
        "frontend\src\hooks",
        "frontend\src\pages",
        "frontend\src\services",
        "frontend\src\utils",
        "frontend\src\styles"
    )
    
    foreach ($dir in $frontendDirs) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }
    
    Write-Success "Project structure created"
} catch {
    Write-Error "Failed to create project structure: $_"
    exit 1
}

# Step 2: Download configuration files from GitHub (if you have a repo)
# For now, we'll inform the user to copy the files manually
Write-Step "Setting Up Configuration Files"
Write-Info "Please copy the following files to their respective locations:"
Write-Info "1. Backend .env file to: $ProjectPath\backend\.env"
Write-Info "2. Frontend .env file to: $ProjectPath\frontend\.env" 
Write-Info "3. go.mod file to: $ProjectPath\backend\go.mod"
Write-Info "4. .air.toml file to: $ProjectPath\backend\.air.toml"
Write-Info "5. Database init scripts to: $ProjectPath\backend\migrations\"

$ready = Read-Host "`nHave you copied all configuration files? (yes/no)"
if ($ready -ne "yes") {
    Write-Warning "Please copy all configuration files before continuing."
    exit 0
}

# Step 3: Initialize Backend
Write-Step "Initializing Go Backend"

Set-Location "$ProjectPath\backend"

# Download Go dependencies
Write-Info "Downloading Go dependencies..."
& go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to download Go dependencies"
    exit 1
}

# Install Air for hot reload
Write-Info "Installing Air for hot reload..."
& go install github.com/cosmtrek/air@latest
if ($LASTEXITCODE -ne 0) {
    Write-Warning "Failed to install Air. You can still run the backend without hot reload."
}

Write-Success "Backend initialized"

# Step 4: Setup MySQL Database
Write-Step "Setting Up MySQL Database"

Write-Info "Running MySQL initialization script..."
Write-Info "You'll be prompted for the MySQL root password (default: root123)"

$mysqlScript = "$ProjectPath\backend\migrations\init-database.sql"
if (Test-Path $mysqlScript) {
    # PowerShell-compatible way to run MySQL script
    $mysqlCommand = "source " + $mysqlScript.Replace('\', '/')
    & mysql -u root -p -e $mysqlCommand
    if ($LASTEXITCODE -eq 0) {
        Write-Success "MySQL database initialized"
    } else {
        Write-Error "Failed to initialize MySQL database"
        Write-Info "You can run the script manually later in MySQL:"
        Write-Info "  1. Open Command Prompt"
        Write-Info "  2. Type: mysql -u root -p"
        Write-Info "  3. Enter password: root123"
        Write-Info "  4. Type: source $($mysqlScript.Replace('\', '/'))"
    }
} else {
    Write-Warning "MySQL init script not found at: $mysqlScript"
}

# Step 5: Setup MongoDB Collections
Write-Step "Setting Up MongoDB Collections"

$mongoScript = "$ProjectPath\backend\migrations\init-mongodb.js"
if (Test-Path $mongoScript) {
    Write-Info "Running MongoDB initialization script..."
    # PowerShell-compatible way to run MongoDB script
    $mongoContent = Get-Content $mongoScript -Raw
    & mongosh --eval $mongoContent
    if ($LASTEXITCODE -eq 0) {
        Write-Success "MongoDB collections created"
    } else {
        Write-Error "Failed to initialize MongoDB"
        Write-Info "You can run the script manually later:"
        Write-Info "  1. Open Command Prompt"
        Write-Info "  2. Navigate to: $ProjectPath\backend\migrations"
        Write-Info "  3. Type: mongosh"
        Write-Info "  4. Copy and paste the contents of init-mongodb.js"
    }
} else {
    Write-Warning "MongoDB init script not found at: $mongoScript"
}

# Step 6: Initialize Frontend
Write-Step "Initializing React Frontend"

Set-Location "$ProjectPath\frontend"

# Create package.json if it doesn't exist
if (-not (Test-Path "package.json")) {
    Write-Info "Creating package.json..."
    $packageJson = @{
        name = "tournament-planner-frontend"
        version = "1.0.0"
        private = $true
        dependencies = @{
            "react" = "^18.2.0"
            "react-dom" = "^18.2.0"
            "react-router-dom" = "^6.20.0"
            "react-query" = "^3.39.3"
            "axios" = "^1.6.2"
            "react-hot-toast" = "^2.4.1"
            "lucide-react" = "^0.294.0"
            "date-fns" = "^2.30.0"
            "@headlessui/react" = "^1.7.17"
            "clsx" = "^2.0.0"
        }
        devDependencies = @{
            "@types/react" = "^18.2.43"
            "@types/react-dom" = "^18.2.17"
            "react-scripts" = "5.0.1"
            "tailwindcss" = "^3.3.6"
            "autoprefixer" = "^10.4.16"
            "postcss" = "^8.4.32"
        }
        scripts = @{
            start = "react-scripts start"
            build = "react-scripts build"
            test = "react-scripts test"
            eject = "react-scripts eject"
        }
        eslintConfig = @{
            extends = @("react-app")
        }
        browserslist = @{
            production = @(">0.2%", "not dead", "not op_mini all")
            development = @("last 1 chrome version", "last 1 firefox version", "last 1 safari version")
        }
    }
    
    $packageJson | ConvertTo-Json -Depth 10 | Set-Content -Path "package.json"
    Write-Success "package.json created"
}

# Install frontend dependencies
Write-Info "Installing frontend dependencies (this may take a few minutes)..."
& npm install
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to install frontend dependencies"
    exit 1
}

# Setup Tailwind CSS
Write-Info "Setting up Tailwind CSS..."
& npx tailwindcss init -p

# Create tailwind.config.js
$tailwindConfig = @"
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
        }
      }
    },
  },
  plugins: [],
}
"@
$tailwindConfig | Set-Content -Path "tailwind.config.js"

Write-Success "Frontend initialized"

# Step 7: Create startup scripts
Write-Step "Creating Startup Scripts"

# Backend startup script
$backendScript = @"
@echo off
echo Starting Tournament Planner Backend...
cd /d "$ProjectPath\backend"
air
pause
"@
$backendScript | Set-Content -Path "$ProjectPath\start-backend.bat"

# Frontend startup script
$frontendScript = @"
@echo off
echo Starting Tournament Planner Frontend...
cd /d "$ProjectPath\frontend"
npm start
pause
"@
$frontendScript | Set-Content -Path "$ProjectPath\start-frontend.bat"

# Combined startup script
$combinedScript = @"
@echo off
echo Starting Tournament Planner...
echo.
echo This will open two windows:
echo 1. Backend server (Go)
echo 2. Frontend server (React)
echo.
start "Tournament Planner Backend" cmd /k "$ProjectPath\start-backend.bat"
timeout /t 5
start "Tournament Planner Frontend" cmd /k "$ProjectPath\start-frontend.bat"
echo.
echo Both servers are starting...
echo Backend will be available at: http://localhost:8080
echo Frontend will be available at: http://localhost:3000
echo.
pause
"@
$combinedScript | Set-Content -Path "$ProjectPath\start-tournament-planner.bat"

Write-Success "Startup scripts created"

# Final Instructions
Write-Host "`n" -NoNewline
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║              Setup Complete! 🎉                                ║" -ForegroundColor Green
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Green

Write-Host "`nYour Tournament Planner is ready to run!" -ForegroundColor Cyan
Write-Host "`nTo start the application:" -ForegroundColor Yellow

Write-Host "1. Double-click: " -NoNewline
Write-Host "$ProjectPath\start-tournament-planner.bat" -ForegroundColor White

Write-Host "`nOr run individually:"
Write-Host "2. Backend: " -NoNewline
Write-Host "$ProjectPath\start-backend.bat" -ForegroundColor White
Write-Host "3. Frontend: " -NoNewline  
Write-Host "$ProjectPath\start-frontend.bat" -ForegroundColor White

Write-Host "`nAccess the application at:" -ForegroundColor Yellow
Write-Host "   Frontend: " -NoNewline
Write-Host "http://localhost:3000" -ForegroundColor White
Write-Host "   Backend API: " -NoNewline
Write-Host "http://localhost:8080" -ForegroundColor White
Write-Host "   Health Check: " -NoNewline
Write-Host "http://localhost:8080/health" -ForegroundColor White

Write-Host "`nDefault Admin Login:" -ForegroundColor Yellow
Write-Host "   Email: " -NoNewline
Write-Host "admin@tournament.local" -ForegroundColor White
Write-Host "   Password: " -NoNewline
Write-Host "admin123" -ForegroundColor White

Write-Host "`nHappy Tournament Planning!" -ForegroundColor Green