# Save this as: simple-setup.ps1

# Set project path
$ProjectPath = "D:\TournamentPlanner"

Write-Host "Tournament Planner Simple Setup" -ForegroundColor Green
Write-Host "===============================" -ForegroundColor Green

# Create all directories
Write-Host "`nCreating project structure..." -ForegroundColor Yellow

$dirs = @(
    "$ProjectPath\backend\cmd\server",
    "$ProjectPath\backend\internal\api",
    "$ProjectPath\backend\internal\config",
    "$ProjectPath\backend\internal\database",
    "$ProjectPath\backend\internal\middleware",
    "$ProjectPath\backend\internal\models",
    "$ProjectPath\backend\internal\repositories",
    "$ProjectPath\backend\internal\services",
    "$ProjectPath\backend\internal\server",
    "$ProjectPath\backend\internal\websocket",
    "$ProjectPath\backend\internal\utils",
    "$ProjectPath\backend\migrations",
    "$ProjectPath\backend\uploads",
    "$ProjectPath\backend\tmp",
    "$ProjectPath\frontend\public",
    "$ProjectPath\frontend\src\components\layout",
    "$ProjectPath\frontend\src\components\ui",
    "$ProjectPath\frontend\src\components\tournaments",
    "$ProjectPath\frontend\src\contexts",
    "$ProjectPath\frontend\src\hooks",
    "$ProjectPath\frontend\src\pages",
    "$ProjectPath\frontend\src\services",
    "$ProjectPath\frontend\src\utils",
    "$ProjectPath\frontend\src\styles"
)

foreach ($dir in $dirs) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
}

Write-Host "[OK] Project structure created" -ForegroundColor Green

# Initialize Go backend
Write-Host "`nInitializing Go backend..." -ForegroundColor Yellow
Set-Location "$ProjectPath\backend"

# Download dependencies
Write-Host "Downloading Go dependencies..." -ForegroundColor Cyan
& go mod download
& go mod tidy

Write-Host "[OK] Backend initialized" -ForegroundColor Green

# Initialize frontend
Write-Host "`nInitializing React frontend..." -ForegroundColor Yellow
Set-Location "$ProjectPath\frontend"

# Check if package.json exists
if (-not (Test-Path "package.json")) {
    Write-Host "Creating package.json..." -ForegroundColor Cyan
    # Create a basic package.json
    $packageContent = @'
{
  "name": "tournament-planner-frontend",
  "version": "1.0.0",
  "private": true,
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "react-query": "^3.39.3",
    "axios": "^1.6.2",
    "react-hot-toast": "^2.4.1",
    "lucide-react": "^0.294.0"
  },
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "test": "react-scripts test"
  },
  "devDependencies": {
    "react-scripts": "5.0.1",
    "tailwindcss": "^3.3.6",
    "autoprefixer": "^10.4.16",
    "postcss": "^8.4.32"
  }
}
'@
    $packageContent | Out-File -FilePath "package.json" -Encoding UTF8
}

Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
& npm install

Write-Host "[OK] Frontend initialized" -ForegroundColor Green

# Create startup batch files
Write-Host "`nCreating startup scripts..." -ForegroundColor Yellow

# Backend startup
@"
@echo off
echo Starting Tournament Planner Backend...
cd /d $ProjectPath\backend
air
pause
"@ | Out-File -FilePath "$ProjectPath\start-backend.bat" -Encoding ASCII

# Frontend startup
@"
@echo off
echo Starting Tournament Planner Frontend...
cd /d $ProjectPath\frontend
npm start
pause
"@ | Out-File -FilePath "$ProjectPath\start-frontend.bat" -Encoding ASCII

# Combined startup
@"
@echo off
echo Starting Tournament Planner...
start "Backend" cmd /k "$ProjectPath\start-backend.bat"
timeout /t 5
start "Frontend" cmd /k "$ProjectPath\start-frontend.bat"
pause
"@ | Out-File -FilePath "$ProjectPath\start-tournament-planner.bat" -Encoding ASCII

Write-Host "[OK] Startup scripts created" -ForegroundColor Green

# Final instructions
Write-Host "`n================================" -ForegroundColor Green
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green
Write-Host "`nNext Steps:" -ForegroundColor Yellow
Write-Host "1. Set up MySQL database manually (see guide)" -ForegroundColor White
Write-Host "2. Set up MongoDB collections manually (see guide)" -ForegroundColor White
Write-Host "3. Make sure all .env files are in place" -ForegroundColor White
Write-Host "4. Run: $ProjectPath\start-tournament-planner.bat" -ForegroundColor White
Write-Host "`nHappy coding!" -ForegroundColor Green