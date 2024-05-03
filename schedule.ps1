# PowerShell script
$command = ".\out\go_build_main_go.exe -localIP '172.18.4.85' -url 'https://www.kuaishou.com/' -parallel 32 -crawlerMode true"
$process = Start-Process -FilePath PowerShell -ArgumentList "-Command $command" -PassThru -WindowStyle Normal

# Wait for 1 hour
Start-Sleep -Seconds 36000

# After 1 hour, if the process is still running, stop it
if (!$process.HasExited) {
    Stop-Process -Id $process.Id
}