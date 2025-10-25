Write-Host "🚀 Starting ChatServer setup for Windows..." -ForegroundColor Cyan

# --- Step 1: Check Go ---
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "⚙️ Go not found. Installing via winget..." -ForegroundColor Yellow
    try {
        winget install -e --id GoLang.Go
    } catch {
        Write-Host "❌ Failed to install Go automatically. Please install it manually from https://go.dev/dl/" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✅ Go is installed" -ForegroundColor Green
}

# --- Step 2: Check Git ---
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host "⚙️ Installing Git via winget..." -ForegroundColor Yellow
    try {
        winget install -e --id Git.Git
    } catch {
        Write-Host "❌ Failed to install Git automatically. Please install it manually." -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✅ Git is installed" -ForegroundColor Green
}

# --- Step 3: Setup and Build ---
Write-Host "🔨 Building ChatServer..." -ForegroundColor Cyan
go mod tidy
if (-not $?) {
    Write-Host "❌ go mod tidy failed" -ForegroundColor Red
    exit 1
}
go build -o chatserver.exe main.go
if (-not $?) {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Build successful" -ForegroundColor Green

# --- Step 4: Kill existing process on port 8080 ---
$port = 8080
$pid = (Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess)
if ($pid) {
    Write-Host "⚙️ Port 8080 in use. Killing process $pid..." -ForegroundColor Yellow
    Stop-Process -Id $pid -Force
}

# --- Step 5: Run the server ---
Write-Host "🗃️ Running ChatServer (close this window to stop)..." -ForegroundColor Cyan
Start-Process -NoNewWindow -FilePath ".\chatserver.exe"

Start-Sleep -Seconds 3
Write-Host "✅ Server running at http://localhost:8080" -ForegroundColor Green
Write-Host "👥 Default users created: tom / jerry" -ForegroundColor Cyan

