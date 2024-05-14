param(
    [Parameter(Mandatory = $true)]
    [string]$scriptPath
)

# 创建一个计划任务触发器，每小时触发一次，从早上 9:30 开始
$trigger = New-ScheduledTaskTrigger -At '9:30' -Once -RepetitionInterval (New-TimeSpan -Hours 1) -RepetitionDuration (New-TimeSpan -Hours 24)
# 创建一个计划任务动作
$action = New-ScheduledTaskAction -Execute 'PowerShell.exe' -Argument $scriptPath

# 创建设置集，设置任务的执行时间限制为 1 小时
$settings = New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Hours 1)

# 创建计划任务
Register-ScheduledTask -TaskName 'HttpBenchmarkScheduledTask' -Trigger $trigger -Action $action -Description 'Runs schedule.ps1 every hour' -Settings $settings