#Requires -Version 5.1
<#
.SYNOPSIS
    启动或重启 Wails 项目。

.DESCRIPTION
    默认行为是"重启"：先关闭旧进程，再启动新实例（单实例规则）。
    若指定 -StartOnly，则仅在项目未运行时启动；已运行时直接通过校验。

.PARAMETER StartOnly
    兼容"启动项目"场景。若项目已在运行，不强制重启，直接校验通过。

.EXAMPLE
    .\scripts\restart-wails.ps1          # 重启项目
    .\scripts\restart-wails.ps1 -StartOnly # 启动项目（已运行则跳过）
#>
[CmdletBinding()]
param(
    [switch]$StartOnly
)

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

# --- 运行态校验 ---

function Test-HttpOk {
    param([string]$Url)
    try {
        $resp = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 2
        return $resp.StatusCode -eq 200
    } catch {
        return $false
    }
}

function Test-ProjectRunning {
    return (Test-HttpOk "http://127.0.0.1:5173") -and (Test-HttpOk "http://localhost:34115")
}

# --- 停止旧进程 ---

function Stop-ProjectProcesses {
    Write-Host "正在关闭旧进程..."

    # 按进程名关闭
    @("wails", "pi-desktop") | ForEach-Object {
        Get-Process -Name $_ -ErrorAction SilentlyContinue | ForEach-Object {
            try {
                Stop-Process -Id $_.Id -Force -ErrorAction SilentlyContinue
                Write-Host "  已关闭进程: $($_.Name) (PID: $($_.Id))"
            } catch {}
        }
    }

    # 按端口关闭残留的 node/vite 进程
    @(5173, 34115) | ForEach-Object {
        $port = $_
        netstat -ano | Select-String ":$port\s+.*LISTENING" | ForEach-Object {
            $parts = ($_.ToString() -split '\s+') | Where-Object { $_ -ne '' }
            if ($parts.Length -ge 5) {
                $pidVal = 0
                if ([int]::TryParse($parts[-1], [ref]$pidVal) -and $pidVal -gt 0 -and $pidVal -ne $PID) {
                    try {
                        Stop-Process -Id $pidVal -Force -ErrorAction SilentlyContinue
                        Write-Host "  已关闭端口 $port 占用进程 (PID: $pidVal)"
                    } catch {}
                }
            }
        }
    }

    Start-Sleep -Seconds 2
}

# --- 启动 ---

function Start-WailsProject {
    param([string]$WailsExe, [string]$WorkingDir, [string]$OutLog, [string]$ErrLog)

    $proc = Start-Process -FilePath $WailsExe `
        -ArgumentList @("dev") `
        -WorkingDirectory $WorkingDir `
        -WindowStyle Hidden `
        -RedirectStandardOutput $OutLog `
        -RedirectStandardError $ErrLog `
        -PassThru

    return $proc
}

function Wait-ForProjectReady {
    param([int]$TimeoutSec = 60)
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    while ($sw.Elapsed.TotalSeconds -lt $TimeoutSec) {
        Start-Sleep -Seconds 1
        if (Test-ProjectRunning) {
            return $true
        }
    }
    return $false
}

# --- 主逻辑 ---

# 1. StartOnly 模式：已在运行则直接通过
if ($StartOnly -and (Test-ProjectRunning)) {
    Write-Host "项目已在运行，跳过启动。"
    Write-Host "  前端 dev server: http://127.0.0.1:5173"
    Write-Host "  Wails 服务:      http://localhost:34115"
    exit 0
}

# 2. 停止旧进程（重启模式，或 StartOnly 但当前未运行时的清理）
Stop-ProjectProcesses

# 3. 清理日志
$outLog = Join-Path $projectRoot "wails-dev.log"
$errLog = Join-Path $projectRoot "wails-dev.err.log"
"" | Set-Content -Path $outLog -ErrorAction SilentlyContinue
"" | Set-Content -Path $errLog -ErrorAction SilentlyContinue

# 4. 查找 wails
$goPath = (go env GOPATH).Trim()
$wails = Join-Path $goPath "bin\wails.exe"
if (-not (Test-Path $wails)) {
    throw "未找到 wails.exe，请先执行：go install github.com/wailsapp/wails/v2/cmd/wails@latest"
}

# 5. 启动
Write-Host "正在启动 wails dev..."
$proc = Start-WailsProject -WailsExe $wails -WorkingDir $projectRoot -OutLog $outLog -ErrLog $errLog

Start-Sleep -Seconds 3

if ($proc.HasExited) {
    $errContent = Get-Content -Path $errLog -Raw -ErrorAction SilentlyContinue
    $outContent = Get-Content -Path $outLog -Raw -ErrorAction SilentlyContinue
    throw @"
wails 进程启动后立即退出。

--- 标准错误 ---
$errContent

--- 标准输出 ---
$outContent
"@
}

# 6. 运行态校验
Write-Host "等待编译完成，最长 60 秒..."
if (-not (Wait-ForProjectReady -TimeoutSec 60)) {
    throw "启动超时：端口 5173 或 34115 未正常响应。请检查日志：$outLog"
}

Write-Host "项目启动成功！"
Write-Host "  前端 dev server: http://127.0.0.1:5173"
Write-Host "  Wails 服务:      http://localhost:34115"
Write-Host "  进程 ID:         $($proc.Id)"
Write-Host "  日志文件:        $outLog"
