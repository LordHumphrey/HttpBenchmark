# Get the task to be copied
$task = Get-ScheduledTask -TaskName "HttpBenchmarkScheduledTask"

# Get the task's action
$action = $task.Actions

# Get the task's trigger
$trigger = $task.Triggers

# Get the task's principal
$principal = New-ScheduledTaskPrincipal -UserId $task.Principal.UserId -LogonType $task.Principal.LogonType

# Register a new task with the same settings
$newTaskName = $task.TaskName + "Nighttime"
Register-ScheduledTask -TaskName $newTaskName -Action $action -Trigger $trigger -Principal $principal