# Tournament Planner Project Setup Script for Windows
# Run this in PowerShell as Administrator

Write-Host "Creating Tournament Planner Project Structure..." -ForegroundColor Green

# Create main project directory
$projectRoot = "C:\TournamentPlanner"
New-Item -ItemType Directory -Force -Path $projectRoot
Set-Location $projectRoot

# Create backend structure
Write-Host "Creating backend structure..." -ForegroundColor Yellow
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
    "backend\pkg"
)

foreach ($dir in $backendDirs) {
    New-Item -ItemType Directory -Force -Path $dir
}

# Create frontend structure
Write-Host "Creating frontend structure..." -ForegroundColor Yellow
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
    New-Item -ItemType Directory -Force -Path $dir
}

Write-Host "Project structure created successfully!" -ForegroundColor Green
Write-Host "Project location: $projectRoot" -ForegroundColor Cyan