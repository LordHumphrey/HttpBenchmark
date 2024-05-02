# PowerShell script
$command = ".\out\go_build_main_go.exe -localIP '172.18.4.85' -url 'https://js.a.kspkg.com/kos/nlav10814/kwai-android-generic-gifmakerrelease-12.3.40.36202_x64_6d5ca4.apk' -parallel 36"
$process = Start-Process -FilePath PowerShell -ArgumentList "-Command $command" -PassThru -WindowStyle Normal

# Wait for 1 hour
Start-Sleep -Seconds 3600

# After 1 hour, if the process is still running, stop it
if (!$process.HasExited) {
    Stop-Process -Id $process.Id
}