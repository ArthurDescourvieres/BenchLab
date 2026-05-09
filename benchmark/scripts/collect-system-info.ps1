# Collecte automatique des conditions de test (machine, OS, versions des outils) - Windows / PowerShell.
# Sortie : benchmark\results\system-info.json + benchmark\results\system-info.txt
$ErrorActionPreference = "Continue"

$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
Set-Location $root
New-Item -ItemType Directory -Force -Path "benchmark\results" | Out-Null

function Get-CmdVersion {
    param([string]$Cmd, [string[]]$ArgList)
    $exe = Get-Command $Cmd -ErrorAction SilentlyContinue
    if ($null -eq $exe) { return "not installed" }
    try {
        $out = & $Cmd @ArgList 2>&1 | Select-Object -First 1
        if ($null -eq $out) { return "unknown" }
        return ($out.ToString()).Trim()
    } catch {
        return "error: $($_.Exception.Message)"
    }
}

$hostnameVal = $env:COMPUTERNAME
if ([string]::IsNullOrEmpty($hostnameVal)) { $hostnameVal = (hostname) }

$os = Get-CimInstance Win32_OperatingSystem
$cpu = Get-CimInstance Win32_Processor | Select-Object -First 1

$osVal = "$($os.Caption) $($os.Version)"
$archVal = $env:PROCESSOR_ARCHITECTURE
$cpuModel = $cpu.Name.Trim()
$cpuCores = $cpu.NumberOfLogicalProcessors
$ramMb = [math]::Round($os.TotalVisibleMemorySize / 1024)

$goVersion = Get-CmdVersion -Cmd "go" -ArgList @("version")
$k6Version = Get-CmdVersion -Cmd "k6" -ArgList @("version")
$ghzVersion = Get-CmdVersion -Cmd "ghz" -ArgList @("--version")
$protocVersion = Get-CmdVersion -Cmd "protoc" -ArgList @("--version")

$timestamp = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$jsonObj = [ordered]@{
    collected_at = $timestamp
    hostname = $hostnameVal
    os = $osVal
    arch = $archVal
    cpu_model = $cpuModel
    cpu_logical_cores = "$cpuCores"
    ram_mb = "$ramMb"
    tools = [ordered]@{
        go = $goVersion
        k6 = $k6Version
        ghz = $ghzVersion
        protoc = $protocVersion
    }
}

$jsonPath = "benchmark\results\system-info.json"
$txtPath = "benchmark\results\system-info.txt"

$jsonObj | ConvertTo-Json -Depth 5 | Out-File -FilePath $jsonPath -Encoding utf8

$txtLines = @()
$txtLines += "BenchLab - Conditions de test (collecte automatique)"
$txtLines += "Date (UTC) : $timestamp"
$txtLines += ""
$txtLines += "[Machine]"
$txtLines += "Hostname            : $hostnameVal"
$txtLines += "OS                  : $osVal"
$txtLines += "Architecture        : $archVal"
$txtLines += "CPU (modele)        : $cpuModel"
$txtLines += "CPU (coeurs logiques): $cpuCores"
$txtLines += "RAM totale          : $ramMb Mo"
$txtLines += ""
$txtLines += "[Outils]"
$txtLines += "Go     : $goVersion"
$txtLines += "k6     : $k6Version"
$txtLines += "ghz    : $ghzVersion"
$txtLines += "protoc : $protocVersion"

$txtLines -join "`r`n" | Out-File -FilePath $txtPath -Encoding utf8

Write-Output "Wrote $jsonPath"
Write-Output "Wrote $txtPath"
