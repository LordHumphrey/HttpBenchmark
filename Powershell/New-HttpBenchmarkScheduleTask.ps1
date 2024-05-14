# Check if the task exists and remove it
if (Get-ScheduledTask -TaskName 'HttpBenchmarkScheduledTaskDaytime' -ErrorAction SilentlyContinue)
{
    Unregister-ScheduledTask -TaskName 'HttpBenchmarkScheduledTaskDaytime' -Confirm:$false
}

# Define the action that the task should perform
$action = New-ScheduledTaskAction -Execute 'PowerShell.exe' -Argument 'H:\Software\RunHttpDownload.ps1 -localIPs (''172.28.0.3'',''172.28.0.3'') -exePath "H:\Software\HttpBenchmark.exe" -processCount 1 -parallelCnt 4' -WorkingDirectory 'H:\Software'

# Define the trigger that will start the task
$trigger = New-ScheduledTaskTrigger -Daily -At 10am

# Register the task with Windows
$task = Register-ScheduledTask -TaskName 'HttpBenchmarkScheduledTaskDaytime' -Action $action -Trigger $trigger -User 'Administrator'

# Modify the task to disallow start if on batteries
$task = Get-ScheduledTask -TaskName 'HttpBenchmarkScheduledTaskDaytime'
$task.Settings.DisallowStartIfOnBatteries = $true
$task.Settings.ExecutionTimeLimit = 'PT14H'
$task | Set-ScheduledTask

# Check if the task exists and remove it
if (Get-ScheduledTask -TaskName 'HttpBenchmarkScheduledTaskNighttime' -ErrorAction SilentlyContinue)
{
    Unregister-ScheduledTask -TaskName 'HttpBenchmarkScheduledTaskNighttime' -Confirm:$false
}

# Define the action that the task should perform
$action = New-ScheduledTaskAction -Execute 'PowerShell.exe' -Argument 'H:\Software\RunHttpDownload.ps1 -localIPs (''172.28.0.3'',''172.28.0.3'') -exePath "H:\Software\HttpBenchmark.exe" -processCount 3 -parallelCnt 8' -WorkingDirectory 'H:\Software'

# Define the trigger that will start the task
$trigger = New-ScheduledTaskTrigger -Daily -At '00:30'

# Register the task with Windows
$task = Register-ScheduledTask -TaskName 'HttpBenchmarkScheduledTaskNighttime' -Action $action -Trigger $trigger -User 'Administrator'

# Modify the task to disallow start if on batteries
$task = Get-ScheduledTask -TaskName 'HttpBenchmarkScheduledTaskNighttime'
$task.Settings.DisallowStartIfOnBatteries = $true
$task.Settings.ExecutionTimeLimit = 'PT9H30M'
$task | Set-ScheduledTask