$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

function Stop-ProjectProcesses {
  Get-Process -Name wails -ErrorAction SilentlyContinue | ForEach-Object { try { Stop-Process -Id $_.Id -Force -ErrorAction SilentlyContinue } catch {} }
  Get-Process -Name pi-desktop-dev -ErrorAction SilentlyContinue | ForEach-Object { try { Stop-Process -Id $_.Id -Force -ErrorAction SilentlyContinue } catch {} }
  Get-Process -Name pi-desktop -ErrorAction SilentlyContinue | ForEach-Object { try { Stop-Process -Id $_.Id -Force -ErrorAction SilentlyContinue } catch {} }

  Get-CimInstance Win32_Process |
    Where-Object {
        $_.Name -eq "node.exe" -and (
        $_.CommandLine -match "pi-app\\frontend" -or
        $_.CommandLine -match "vite" -or
        $_.CommandLine -match 'pnpm\.cjs" run dev'
      )
    } |
    ForEach-Object {
      Stop-Process -Id $_.ProcessId -Force
    }

  Start-Sleep -Seconds 1
}

function Stop-PortOwners {
  param(
    [int[]]$Ports
  )

  foreach ($port in $Ports) {
    $lines = netstat -ano | Select-String ":$port"
    foreach ($line in $lines) {
      $parts = ($line.ToString() -split '\s+') | Where-Object { $_ -ne '' }
      if ($parts.Length -ge 5) {
        $procId = 0
        if ([int]::TryParse($parts[-1], [ref]$procId) -and $procId -gt 0) {
          try { Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue } catch {}
        }
      }
    }
  }
  Start-Sleep -Seconds 1
}

function Wait-ForPort5173 {
  for ($i = 0; $i -lt 40; $i++) {
    Start-Sleep -Seconds 1
    $listening = netstat -ano | Select-String ":5173" | Select-String "LISTENING"
    if ($listening) {
      return $true
    }
  }
  return $false
}

function Wait-ForHttpReady {
  param(
    [string]$Url
  )
  for ($i = 0; $i -lt 40; $i++) {
    Start-Sleep -Seconds 1
    try {
      $resp = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 2
      if ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 500) {
        return $true
      }
    } catch {}
  }
  return $false
}

$wails = Join-Path (go env GOPATH) "bin\wails.exe"
if (-not (Test-Path $wails)) {
  throw "wails.exe not found. Run: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
}

Stop-ProjectProcesses
Stop-PortOwners -Ports @(5173, 34115, 34116)

$outLog = Join-Path $projectRoot "wails-dev.log"
$errLog = Join-Path $projectRoot "wails-dev.err.log"
Set-Content -Path $outLog -Value ""
Set-Content -Path $errLog -Value ""

Start-Process -FilePath $wails `
  -ArgumentList @("dev", "-devserver", "127.0.0.1:34116") `
  -WorkingDirectory $projectRoot `
  -WindowStyle Hidden `
  -RedirectStandardOutput $outLog `
  -RedirectStandardError $errLog

Start-Sleep -Seconds 2
$current = Get-Process -Name wails -ErrorAction SilentlyContinue
if (-not $current) {
  throw "wails failed to start."
}

if (-not (Wait-ForPort5173)) {
  throw "startup failed: port 5173 is not listening."
}
if (-not (Wait-ForHttpReady -Url "http://127.0.0.1:5173")) {
  throw "startup failed: frontend dev server is not responding."
}

Write-Host "Wails restarted in background mode. Current instance count: $($current.Count)"
$current | Select-Object Id, ProcessName, StartTime
