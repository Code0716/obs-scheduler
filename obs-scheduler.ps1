<#
.SYNOPSIS
    OBS録画スケジューラー (PowerShell版)
    Smart App Control 対策として powershell.exe 経由で動作させるバージョン。

.PARAMETER Start
    録画開始時刻 (HH:mm 形式、例: "19:00")

.PARAMETER Stop
    録画停止時刻 (HH:mm 形式、例: "21:00")

.PARAMETER OBSAddr
    OBS WebSocket のアドレス (デフォルト: ws://localhost:4455)

.PARAMETER OBSPassword
    OBS WebSocket のパスワード

.PARAMETER OBSPath
    OBS実行ファイルのパス (省略時はデフォルトパスを使用)

.PARAMETER SkipLaunch
    OBSの自動起動をスキップするフラグ

.EXAMPLE
    .\obs-scheduler.ps1 -Start "19:00" -Stop "21:00"
    .\obs-scheduler.ps1 -Start "19:00" -Stop "21:00" -OBSPassword "mypassword"
    .\obs-scheduler.ps1 -Start "19:00" -Stop "21:00" -SkipLaunch
#>
param(
    [string]$Start      = "08:44",
    [string]$Stop       = "10:00",
    [string]$OBSAddr    = "ws://localhost:4455",
    [string]$OBSPassword = $env:OBS_PASSWORD,
    [string]$OBSPath    = $env:OBS_APP_PATH,
    [switch]$SkipLaunch
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# ─────────────────────────────────────────────
# ログ出力ヘルパー
# ─────────────────────────────────────────────
function Write-Log {
    param([string]$Level, [string]$Message, [hashtable]$Fields = @{})
    $timestamp = Get-Date -Format "yyyy-MM-ddTHH:mm:ss"
    $extra = if ($Fields.Count -gt 0) { ($Fields.GetEnumerator() | ForEach-Object { "$($_.Key)=$($_.Value)" }) -join " " } else { "" }
    Write-Host "[$timestamp] [$Level] $Message $extra".TrimEnd()
}

# ─────────────────────────────────────────────
# 指定時刻まで待機する
# ─────────────────────────────────────────────
function Wait-Until {
    param([DateTime]$TargetTime, [string]$Label)

    $now = Get-Date
    if ($TargetTime -lt $now) {
        Write-Log "WARN" "$Label はすでに過去の時刻です。スキップします。" @{ target = $TargetTime.ToString("HH:mm:ss") }
        return
    }

    $wait = $TargetTime - $now
    Write-Log "INFO" "$Label まで待機します" @{ target = $TargetTime.ToString("HH:mm:ss"); wait_sec = [int]$wait.TotalSeconds }

    # 1秒ごとにCtrl+Cを受け付けながら待機
    $until = $TargetTime
    while ((Get-Date) -lt $until) {
        Start-Sleep -Seconds 1
    }
}

# ─────────────────────────────────────────────
# OBS WebSocket v5 クライアント
# ─────────────────────────────────────────────
function New-OBSWebSocketClient {
    param([string]$Uri, [string]$Password)

    # .NET ClientWebSocket を使用 (PowerShell 5.1+ / .NET 4.5+)
    $ws = [System.Net.WebSockets.ClientWebSocket]::new()
    $ct = [System.Threading.CancellationToken]::None

    Write-Log "INFO" "OBS WebSocket に接続中..." @{ uri = $Uri }
    try {
        $ws.ConnectAsync([Uri]$Uri, $ct).GetAwaiter().GetResult()
    }
    catch {
        throw "OBS WebSocket への接続に失敗しました: $_"
    }

    # ── Hello 受信 (Op=0) ──
    $recvBuf = [byte[]]::new(8192)
    $segment = [System.ArraySegment[byte]]::new($recvBuf)
    $result  = $ws.ReceiveAsync($segment, $ct).GetAwaiter().GetResult()
    $helloJson = [System.Text.Encoding]::UTF8.GetString($recvBuf, 0, $result.Count)
    $hello = $helloJson | ConvertFrom-Json

    Write-Log "DEBUG" "Hello 受信" @{ op = $hello.op }

    # ── 認証ハッシュ生成 ──
    $sha256 = [System.Security.Cryptography.SHA256]::Create()
    $salt      = $hello.d.authentication.salt
    $challenge = $hello.d.authentication.challenge

    $secretHash = $sha256.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($Password + $salt))
    $secret     = [Convert]::ToBase64String($secretHash)
    $authHash   = $sha256.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($secret + $challenge))
    $auth       = [Convert]::ToBase64String($authHash)
    $sha256.Dispose()

    # ── Identify 送信 (Op=1) ──
    $identify = [ordered]@{
        op = 1
        d  = [ordered]@{
            rpcVersion     = $hello.d.rpcVersion
            authentication = $auth
        }
    } | ConvertTo-Json -Compress

    $sendBytes   = [System.Text.Encoding]::UTF8.GetBytes($identify)
    $sendSegment = [System.ArraySegment[byte]]::new($sendBytes)
    $ws.SendAsync($sendSegment, [System.Net.WebSockets.WebSocketMessageType]::Text, $true, $ct).GetAwaiter().GetResult()

    # ── Identified 受信 (Op=2) ──
    $result2  = $ws.ReceiveAsync($segment, $ct).GetAwaiter().GetResult()
    $identifiedJson = [System.Text.Encoding]::UTF8.GetString($recvBuf, 0, $result2.Count)
    $identified = $identifiedJson | ConvertFrom-Json
    if ($identified.op -ne 2) {
        throw "予期しない OpCode を受信しました: $($identified.op). 認証に失敗した可能性があります。"
    }

    Write-Log "INFO" "OBS WebSocket に接続・認証しました"
    return $ws
}

function Send-OBSRequest {
    param(
        [System.Net.WebSockets.ClientWebSocket]$WebSocket,
        [string]$RequestType
    )

    $ct = [System.Threading.CancellationToken]::None
    $requestId = [System.Guid]::NewGuid().ToString()

    $req = [ordered]@{
        op = 6
        d  = [ordered]@{
            requestType = $RequestType
            requestId   = $requestId
        }
    } | ConvertTo-Json -Compress

    Write-Log "INFO" "OBS リクエスト送信" @{ requestType = $RequestType }

    $sendBytes   = [System.Text.Encoding]::UTF8.GetBytes($req)
    $sendSegment = [System.ArraySegment[byte]]::new($sendBytes)
    $WebSocket.SendAsync($sendSegment, [System.Net.WebSockets.WebSocketMessageType]::Text, $true, $ct).GetAwaiter().GetResult()

    # レスポンス待機 (Op=7 が来るまでループ)
    $recvBuf = [byte[]]::new(8192)
    $segment = [System.ArraySegment[byte]]::new($recvBuf)

    $deadline = [DateTime]::Now.AddSeconds(30)
    while ([DateTime]::Now -lt $deadline) {
        $result = $WebSocket.ReceiveAsync($segment, $ct).GetAwaiter().GetResult()
        $respJson = [System.Text.Encoding]::UTF8.GetString($recvBuf, 0, $result.Count)
        $resp = $respJson | ConvertFrom-Json

        if ($resp.op -eq 7 -and $resp.d.requestId -eq $requestId) {
            if (-not $resp.d.requestStatus.result) {
                throw "OBS リクエスト失敗: code=$($resp.d.requestStatus.code) comment=$($resp.d.requestStatus.comment)"
            }
            Write-Log "INFO" "OBS リクエスト成功" @{ requestType = $RequestType }
            return
        }
        # Op=5 はイベント通知。StopRecord後の RecordStateChanged を待機
        if ($resp.op -eq 5 -and $RequestType -eq "StopRecord") {
            if ($resp.d.eventType -eq "RecordStateChanged" -and
                $resp.d.eventData.outputState -eq "OBS_WEBSOCKET_OUTPUT_STOPPED") {
                Write-Log "INFO" "録画停止イベントを確認しました"
                return
            }
        }
    }
    throw "OBS レスポンスがタイムアウトしました (30秒)"
}

function Close-OBSWebSocket {
    param([System.Net.WebSockets.ClientWebSocket]$WebSocket)
    $ct = [System.Threading.CancellationToken]::None
    try {
        $WebSocket.CloseAsync(
            [System.Net.WebSockets.WebSocketCloseStatus]::NormalClosure,
            "Done", $ct
        ).GetAwaiter().GetResult()
    }
    catch { <# 接続が既に切れている場合は無視 #> }
    $WebSocket.Dispose()
}

# ─────────────────────────────────────────────
# OBS の起動・終了
# ─────────────────────────────────────────────
function Start-OBSApp {
    param([string]$Path)

    if (-not $Path) {
        $Path = "C:\Program Files\obs-studio\bin\64bit\obs64.exe"
    }

    if (-not (Test-Path $Path)) {
        throw "OBS が見つかりません: $Path"
    }

    Write-Log "INFO" "OBS を起動します" @{ path = $Path }
    Start-Process -FilePath $Path
}

function Stop-OBSApp {
    $procs = Get-Process -Name "obs64" -ErrorAction SilentlyContinue
    if ($procs) {
        Write-Log "INFO" "OBS を終了します"
        $procs | Stop-Process -ErrorAction SilentlyContinue
    }
}

# ─────────────────────────────────────────────
# リトライ付き接続
# ─────────────────────────────────────────────
function Connect-OBSWithRetry {
    param([string]$Uri, [string]$Password, [int]$MaxRetry = 15, [int]$IntervalSec = 2)

    for ($i = 1; $i -le $MaxRetry; $i++) {
        try {
            return New-OBSWebSocketClient -Uri $Uri -Password $Password
        }
        catch {
            Write-Log "WARN" "OBS への接続試行 $i/$MaxRetry 失敗。リトライします..." @{ error = $_.Exception.Message }
            if ($i -lt $MaxRetry) { Start-Sleep -Seconds $IntervalSec }
        }
    }
    throw "OBS WebSocket への接続が $MaxRetry 回失敗しました"
}

# ─────────────────────────────────────────────
# メイン処理
# ─────────────────────────────────────────────
function Main {
    $today     = (Get-Date).Date
    $startTime = $today.Add([TimeSpan]::Parse($Start))
    $stopTime  = $today.Add([TimeSpan]::Parse($Stop))

    Write-Log "INFO" "OBS スケジューラー 開始" @{
        start    = $startTime.ToString("HH:mm")
        stop     = $stopTime.ToString("HH:mm")
        obs_addr = $OBSAddr
    }

    # OBS 起動 (開始時刻の10秒前)
    if (-not $SkipLaunch) {
        $launchTime = $startTime.AddSeconds(-10)
        Wait-Until -TargetTime $launchTime -Label "OBS 起動"
        Start-OBSApp -Path $OBSPath
    }

    # OBS WebSocket 接続 (リトライあり)
    $ws = Connect-OBSWithRetry -Uri $OBSAddr -Password $OBSPassword

    try {
        # 録画開始時刻まで待機
        Wait-Until -TargetTime $startTime -Label "録画開始"
        Send-OBSRequest -WebSocket $ws -RequestType "StartRecord"
        Write-Log "INFO" "録画を開始しました"

        # 録画停止時刻まで待機
        Wait-Until -TargetTime $stopTime -Label "録画停止"
        Send-OBSRequest -WebSocket $ws -RequestType "StopRecord"
        Write-Log "INFO" "録画を停止しました"
    }
    finally {
        Close-OBSWebSocket -WebSocket $ws
    }

    # OBS 終了
    if (-not $SkipLaunch) {
        Start-Sleep -Seconds 2
        Stop-OBSApp
    }

    Write-Log "INFO" "スケジューラー 正常終了"
}

# ─────────────────────────────────────────────
# エントリポイント
# ─────────────────────────────────────────────
try {
    Main
}
catch {
    Write-Log "ERROR" "致命的エラーが発生しました" @{ error = $_.Exception.Message }
    exit 1
}
