[CmdletBinding()]
param(
    [switch]$Race,
    [switch]$Vet,
    [switch]$Coverage
)

$ErrorActionPreference = "Stop"

function Initialize-Utf8Console {
    $utf8NoBom = [System.Text.UTF8Encoding]::new($false)
    [Console]::InputEncoding = $utf8NoBom
    [Console]::OutputEncoding = $utf8NoBom
    $script:OutputEncoding = $utf8NoBom

    $PSDefaultParameterValues["Get-Content:Encoding"] = "UTF8"
    $PSDefaultParameterValues["Select-String:Encoding"] = "UTF8"
}

function Test-IsWindows {
    $isWindowsVariable = Get-Variable -Name IsWindows -ErrorAction SilentlyContinue
    if ($null -ne $isWindowsVariable) {
        return [bool]$isWindowsVariable.Value
    }

    return [System.Environment]::OSVersion.Platform -eq [System.PlatformID]::Win32NT
}

function Set-DefaultEnvPath {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    $current = [System.Environment]::GetEnvironmentVariable($Name, "Process")
    if (-not [string]::IsNullOrWhiteSpace($current)) {
        Write-Host "Using $Name=$current"
        return
    }

    New-Item -ItemType Directory -Force -Path $Path | Out-Null
    [System.Environment]::SetEnvironmentVariable($Name, $Path, "Process")
    Write-Host "Using $Name=$Path"
}

function Initialize-GoWorkspacePaths {
    param(
        [Parameter(Mandatory = $true)]
        [string]$ProjectRoot
    )

    if (-not (Test-IsWindows)) {
        return
    }

    Set-DefaultEnvPath -Name "GOCACHE" -Path (Join-Path $ProjectRoot ".gocache")
    Set-DefaultEnvPath -Name "GOTMPDIR" -Path (Join-Path $ProjectRoot ".gotmp")
}

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        [scriptblock]$Step
    )

    Write-Host "==> $Name"
    & $Step
    if ($LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $LASTEXITCODE"
    }
}

function Resolve-GoCommand {
    $goCommand = Get-Command go -ErrorAction SilentlyContinue
    if ($goCommand) {
        return $goCommand.Source
    }

    $defaultWindowsGo = "C:\Program Files\Go\bin\go.exe"
    if (Test-Path $defaultWindowsGo) {
        return $defaultWindowsGo
    }

    throw "Go toolchain not found. Install Go 1.22 or newer, add it to PATH, or install it at $defaultWindowsGo."
}

$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    Initialize-Utf8Console
    Initialize-GoWorkspacePaths -ProjectRoot $projectRoot

    $go = Resolve-GoCommand

    $goVersion = & $go env GOVERSION
    if ($LASTEXITCODE -ne 0) {
        throw "failed to read Go version"
    }
    Write-Host "Using $goVersion"

    Invoke-Step "Download modules" { & $go mod download }

    if ($Coverage) {
        Invoke-Step "Run tests with coverage" { & $go test ./... -coverprofile coverage.out }
        Invoke-Step "Show coverage summary" { & $go tool cover -func coverage.out }
    } else {
        Invoke-Step "Run tests" { & $go test ./... }
    }

    if ($Race) {
        Invoke-Step "Run race tests" { & $go test -race ./... }
    }

    if ($Vet) {
        Invoke-Step "Run go vet" { & $go vet ./... }
    }
} finally {
    Pop-Location
}
