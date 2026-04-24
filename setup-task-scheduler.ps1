<#
.SYNOPSIS
    Windowsタスクスケジューラに obs-scheduler.ps1 を登録するセットアップスクリプト。
    管理者権限で実行してください。

.PARAMETER Start
    録画開始時刻 (HH:mm 形式)

.PARAMETER Stop
    録画停止時刻 (HH:mm 形式)

.PARAMETER ScriptDir
    obs-scheduler.ps1 があるディレクトリ (省略時はこのスクリプトと同じ場所)

.PARAMETER OBSPassword
    OBS WebSocket パスワード

.PARAMETER TaskName
    タスクスケジューラに登録するタスク名 (デフォルト: OBS-Scheduler)

.PARAMETER Remove
    タスクを削除するフラグ

.EXAMPLE
    # 管理者権限のPowerShellで実行
    .\setup-task-scheduler.ps1 -Start "19:00" -Stop "21:00" -OBSPassword "mypass"
    .\setup-task-scheduler.ps1 -Remove
#>
param(
    [string]$Start       = "08:44",
    [string]$Stop        = "10:00",
    [string]$ScriptDir   = $PSScriptRoot,
    [string]$OBSPassword = "",
    [string]$TaskName    = "OBS-Scheduler",
    [switch]$Remove
)

$schedulerScript = Join-Path $ScriptDir "obs-scheduler.ps1"

# ── タスク削除 ──
if ($Remove) {
    if (Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue) {
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
        Write-Host "タスク '$TaskName' を削除しました。"
    } else {
        Write-Host "タスク '$TaskName' は存在しません。"
    }
    exit 0
}

# ── スクリプト存在確認 ──
if (-not (Test-Path $schedulerScript)) {
    Write-Error "obs-scheduler.ps1 が見つかりません: $schedulerScript"
    exit 1
}

# ── 引数文字列組み立て ──
$args = "-NonInteractive -ExecutionPolicy Bypass -File `"$schedulerScript`" -Start `"$Start`" -Stop `"$Stop`""
if ($OBSPassword) {
    $args += " -OBSPassword `"$OBSPassword`""
}

# ── 開始時刻を今日の日付で計算 ──
$triggerTime = [DateTime]::Today.Add([TimeSpan]::Parse($Start)).AddSeconds(-15)
if ($triggerTime -lt (Get-Date)) {
    # 今日の時刻が過ぎていれば翌日に設定
    $triggerTime = $triggerTime.AddDays(1)
}

# ── タスク定義 ──
$action  = New-ScheduledTaskAction -Execute "powershell.exe" -Argument $args
$trigger = New-ScheduledTaskTrigger -Once -At $triggerTime `
           -RepetitionInterval (New-TimeSpan -Days 1) `
           -RepetitionDuration ([TimeSpan]::MaxValue)

$settings = New-ScheduledTaskSettingsSet `
    -ExecutionTimeLimit (New-TimeSpan -Hours 6) `
    -MultipleInstances IgnoreNew `
    -StartWhenAvailable

$principal = New-ScheduledTaskPrincipal `
    -UserId ([System.Security.Principal.WindowsIdentity]::GetCurrent().Name) `
    -LogonType Interactive `
    -RunLevel Highest

# ── 登録 (既存があれば上書き) ──
$task = New-ScheduledTask -Action $action -Trigger $trigger -Settings $settings -Principal $principal

Register-ScheduledTask -TaskName $TaskName -InputObject $task -Force | Out-Null

Write-Host ""
Write-Host "✅ タスクスケジューラへの登録が完了しました" -ForegroundColor Green
Write-Host "   タスク名     : $TaskName"
Write-Host "   録画開始     : $Start"
Write-Host "   録画停止     : $Stop"
Write-Host "   初回実行予定 : $($triggerTime.ToString('yyyy-MM-dd HH:mm:ss'))"
Write-Host "   以降         : 毎日同時刻に自動実行"
Write-Host ""
Write-Host "タスクスケジューラで確認: taskschd.msc"
