# PowerShell script
param (
    $localIPs = @('172.18.4.85', '172.18.4.90'), # Default IPs
    $exePath = 'F:\Golang\HttpBenchmark\out\HttpBenchmark.exe', # Default executable path
    $processCount = 1, # Default process count
    $parallelCnt = 8,
    $source = 'winget', # Default source
    $name = @('Baidu', 'NetEase', 'Youku', 'iQIYI', 'Tencent', 'Bilibili', 'Sohu', 'Xunlei', 'Douyu', 'Alibaba', 'Xiaomi', 'JetBrains', 'Microsoft', 'Intel')
)
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
# H:\Software\RunHttpDownload.ps1 -localIPs ('172.28.0.3','172.28.0.3') -exePath "H:\Software\HttpBenchmark.exe" -processCount 2

function StopExistingProcesses($exeName) {
    # Get all processes with the same name as the executable
    $existingProcesses = Get-Process | Where-Object { $_.Name -eq $exeName }

    # Stop all existing processes
    foreach ($process in $existingProcesses) {
        Write-Host "Stopping existing process $( $process.Id )"
        Stop-Process -Id $process.Id
    }
}

function StartNewProcesses($exePath, $urls, $localIPs, $processCount, $parallelCnt, $crawlerMode) {
    $processes = @()
    $localIPsCount = $localIPs.Length
    $urlsCount = $urls.Length
    for ($i = 0; $i -lt $processCount; $i++) {
        $localIP = $localIPs[$( $i % $localIPsCount )]
        $url = @($urls)[$( $i % $urlsCount )]
        if ($url -ne '') {
            $command = "$exePath  -localIP '$localIP' -url '$url' -parallel $parallelCnt"
            Write-Host "  $( $i % $localIPsCount ),Local IP: $localIP, URL: $url"
            if ($crawlerMode) {
                $command += ' -crawlerMode true'
            }
            try {
                $process = Start-Process -FilePath PowerShell -ArgumentList "-Command $command" -PassThru -NoNewWindow
                $processes += $process
            }
            catch {
                # If the process exits with an error, restart it
                Write-Host $_.Exception.Message
                $process = Start-Process -FilePath PowerShell -ArgumentList "-Command $command" -PassThru -NoNewWindow
                $processes += $process
            }
        }
        else {
            Write-Host 'URL is empty, skipping'
            continue
        }
    }
    return $processes
}

function StopNewProcesses($processes) {
    Write-Host 'Stopping all processes'

    Stop-ScheduledTask -TaskName 'HttpBenchmarkScheduledTask'

    # Stop all processes
    foreach ($process in $processes) {
        if (!$process.HasExited) {
            Stop-Process -Id $process.Id
        }
    }
}

function Get-Urls {
    param (
        [Parameter(Mandatory = $true)]
        [string] $source,

        [Parameter(Mandatory = $false)]
        [string] $id,

        [Parameter(Mandatory = $true)]
        [int] $urlCount
    )

    $urlPairs = @()
    # If $id is not provided, get a random id
    if (-not $id) {
        for ($i = 0; $i -lt $urlCount; $i++) {
            $randomName = Get-Random -InputObject $name
            $idVersionPair = Get-RandomIdandVersionPairs -name $randomName -count 1 -source $source
            $id = $idVersionPair.Id
            $version = $idVersionPair.Version
            # Get version info
            $versionInfo = Invoke-Expression ".\winget-cli_x64\winget.exe show --id $id --exact --disable-interactivity --accept-source-agreements --source $source --version $version"

            $homepage = $null
            $installerUrl = $null

            # Split the version info into lines
            $versionInfoLines = $versionInfo -split "`n"

            # Loop through each line
            foreach ($line in $versionInfoLines) {
                # Check if the line contains 'Homepage:'
                if ($line -match 'Homepage:') {
                    # Get the URL by replacing 'Homepage: ' with nothing
                    $homepage = $line -replace '^Homepage:\s*', ''
                    $homepage = $homepage.Trim() # Remove leading and trailing spaces
                }
                # Check if the line contains 'Installer Url:'
                elseif ($line -match 'Installer Url') {
                    # Get the URL by replacing 'Installer Url: ' with nothing
                    $installerUrl = $line -replace 'Installer Url:', '' # Remove 'Installer Url:'
                    $installerUrl = $installerUrl.Trim() # Remove leading and trailing spaces
                }
            }
            if ($installerUrl -like '*github.com*' -or $installerUrl -like '*githubusercontent.com*') {
                $i--
            }
            else {
                $urlPairs += $installerUrl
            }
        }
    }
    else {
        # Get all available versions
        $allLines = Invoke-Expression -Command ".\winget-cli_x64\winget.exe show --id $id --exact --disable-interactivity --accept-source-agreements --source $source --versions"
        $allLines = $allLines -split "`n"

        $versions = @()
        $foundDash = $false

        foreach ($line in $allLines) {
            if ($foundDash) {
                $versions += $line
            }
            elseif ($line -match '\-') {
                $foundDash = $true
            }
        }


        for ($i = 0; $i -lt $urlCount; $i++) {
            # Get random version
            $randomVersion = Get-Random -InputObject $versions

            # Get version info
            $versionInfo = Invoke-Expression ".\winget-cli_x64\winget.exe show --id $id --exact --disable-interactivity --accept-source-agreements --source $source --version $randomVersion"

            $homepage = $null
            $installerUrl = $null

            # Split the version info into lines
            $versionInfoLines = $versionInfo -split "`n"

            # Loop through each line
            foreach ($line in $versionInfoLines) {
                # Check if the line contains 'Homepage:'
                if ($line -match 'Homepage:') {
                    # Get the URL by replacing 'Homepage: ' with nothing
                    $homepage = $line -replace '^Homepage:\s*', ''
                    $homepage = $homepage.Trim() # Remove leading and trailing spaces
                }
                # Check if the line contains 'Installer Url:'
                elseif ($line -match 'Installer Url') {
                    # Get the URL by replacing 'Installer Url: ' with nothing
                    $installerUrl = $line -replace 'Installer Url:', '' # Remove 'Installer Url:'
                    $installerUrl = $installerUrl.Trim() # Remove leading and trailing spaces
                }
            }
            if ($installerUrl -like '*github.com*' -or $installerUrl -like '*githubusercontent.com*') {
                $i--
            }
            else {
                $urlPairs += $installerUrl
            }
        }
    }
    # Return the requested number of url pairs
    return $urlPairs
}
function Get-RandomIdandVersionPairs {
    param (
        [Parameter(Mandatory = $true)]
        [string] $name,

        [Parameter(Mandatory = $true)]
        [int] $count,

        [Parameter(Mandatory = $true)]
        [string] $source
    )

    # Run the winget search command and capture the output
    $output = .\winget-cli_x64\winget.exe search $name --disable-interactivity --accept-source-agreements --source $source

    # Split the output into lines
    $allLines = $output -split "`n"

    $idVersionPairs = @()
    $foundDash = $false

    foreach ($line in $allLines) {
        if ($foundDash) {
            # Split the line by tab and select the second (id) and fourth (version) columns
            $lineParts = $line -split ' ' | Where-Object { $_.Trim() -ne '' }
            $id = $lineParts[1]
            $version = $lineParts[2]

            # Create a custom object with id and version properties and add it to the array
            $idVersionPairs += New-Object PSObject -Property @{
                Id      = $id
                Version = $version
            }
        }
        elseif ($line -like '--*') {
            $foundDash = $true
        }
    }

    # Randomly select $count Id-Version pairs
    $randomIdVersionPairs = Get-Random -InputObject $idVersionPairs -Count $count

    return $randomIdVersionPairs
}

# Get the name of the executable
$exeName = (Split-Path -Path $exePath -Leaf) -replace '.exe$'

StopExistingProcesses $exeName

$retryCount = 0
do {
    $urls = Get-Urls -source $source -urlCount $processCount
    Write-Host $urls
    $retryCount++
} while (($null -eq $urls -or $urls.Count -ne $processCount) -and $retryCount -le 10)

$allProcesses = @()

$processes = StartNewProcesses $exePath $urls $localIPs $processCount $parallelCnt $crawlerMode
$allProcesses += $processes


# Wait for 1 hour
Start-Sleep -Seconds 3580

StopNewProcesses $allProcesses

