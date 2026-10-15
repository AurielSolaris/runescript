<#
.SYNOPSIS
    Installs RuneScript on Windows.

.DESCRIPTION
    Builds the runescript CLI (compiler and runtime) with your Go toolchain,
    installs it in %LOCALAPPDATA%\Programs\RuneScript\bin, and adds that
    directory to your user PATH. No administrator rights are needed.

    From a RuneScript checkout it builds that source:
        powershell -ExecutionPolicy Bypass -File install.ps1

    Anywhere else it builds the latest version from GitHub:
        irm https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.ps1 | iex

    The installer needs Go 1.25 or newer (https://go.dev/dl/). It uses the
    go on PATH, the one named by RUNESCRIPT_GO, or a standard location such
    as C:\Program Files\Go\bin\go.exe.

.PARAMETER InstallDir
    Install into InstallDir\bin. Default: $env:RUNESCRIPT_HOME, else
    %LOCALAPPDATA%\Programs\RuneScript.

.PARAMETER Version
    Version to build from GitHub: a tag, branch, or commit. Default: latest
    (or $env:RUNESCRIPT_VERSION).

.PARAMETER Remote
    Build from GitHub even when run from a checkout.

.PARAMETER NoModifyPath
    Don't change the user PATH or RUNESCRIPT_HOME.

.PARAMETER Uninstall
    Remove RuneScript and its PATH entry.
#>
[CmdletBinding()]
param(
    [string]$InstallDir = $(if ($env:RUNESCRIPT_HOME) { $env:RUNESCRIPT_HOME } else { Join-Path $env:LOCALAPPDATA 'Programs\RuneScript' }),
    [string]$Version = $(if ($env:RUNESCRIPT_VERSION) { $env:RUNESCRIPT_VERSION } else { 'latest' }),
    [switch]$Remote,
    [switch]$NoModifyPath,
    [switch]$Uninstall,
    # Where the user environment lives. Tests point this at a scratch key.
    [Parameter(DontShow)][string]$EnvKey = 'HKCU:\Environment'
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3

$Module = 'github.com/AurielSolaris/runescript'
$MinGo = [version]'1.25' # keep in sync with go.mod
$MinGoText = '1.25'

function Step([string]$m) { Write-Host "==> $m" -ForegroundColor Cyan }
function Ok([string]$m) { Write-Host "  ok $m" -ForegroundColor Green }
function Warn([string]$m) { Write-Host "warning: $m" -ForegroundColor Yellow }
function Fail([string]$m) {
    Write-Host "error: $m" -ForegroundColor Red
    # exit ends a script run with -File; throw ends one run through iex
    # without closing the user's shell.
    if ($MyInvocation.ScriptName -or $PSCommandPath) { exit 1 }
    throw 'RuneScript installation failed.'
}

$InstallDir = [IO.Path]::GetFullPath($InstallDir)
$BinDir = Join-Path $InstallDir 'bin'
$Target = Join-Path $BinDir 'runescript.exe'

# ---- user environment ----

function Get-RawPath {
    # Read Path without expanding %VARIABLES%, so writing it back keeps them.
    $key = Get-Item -LiteralPath $EnvKey -ErrorAction SilentlyContinue
    if (-not $key) { return '' }
    $v = $key.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
    return [string]$v
}

function Set-RawPath([string]$value) {
    if (-not (Test-Path -LiteralPath $EnvKey)) { New-Item -Path $EnvKey -Force | Out-Null }
    Set-ItemProperty -LiteralPath $EnvKey -Name 'Path' -Value $value -Type ExpandString
}

function Same-Dir([string]$a, [string]$b) {
    $x = [Environment]::ExpandEnvironmentVariables($a).TrimEnd('\')
    $y = [Environment]::ExpandEnvironmentVariables($b).TrimEnd('\')
    return [string]::Equals($x, $y, [StringComparison]::OrdinalIgnoreCase)
}

function Send-SettingChange {
    # Tell Explorer and other programs that the environment changed, so new
    # terminals see it without signing out.
    if (-not ('RuneScript.NativeMethods' -as [type])) {
        Add-Type -Namespace RuneScript -Name NativeMethods -MemberDefinition @'
[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
'@
    }
    $result = [UIntPtr]::Zero
    [void][RuneScript.NativeMethods]::SendMessageTimeout([IntPtr]0xffff, 0x1A, [UIntPtr]::Zero, 'Environment', 2, 5000, [ref]$result)
}

# ---- uninstall ----

if ($Uninstall) {
    Step "Uninstalling RuneScript from $InstallDir"
    if (Test-Path -LiteralPath $Target) { Remove-Item -LiteralPath $Target -Force }
    foreach ($d in @($BinDir, $InstallDir)) {
        if ((Test-Path -LiteralPath $d) -and -not (Get-ChildItem -LiteralPath $d -Force)) { Remove-Item -LiteralPath $d -Force }
    }
    if (Test-Path -LiteralPath $InstallDir) { Warn "$InstallDir is not empty; left in place" }
    $parts = @((Get-RawPath) -split ';' | Where-Object { $_ -and -not (Same-Dir $_ $BinDir) })
    Set-RawPath ($parts -join ';')
    if (Test-Path -LiteralPath $EnvKey) { Remove-ItemProperty -LiteralPath $EnvKey -Name 'RUNESCRIPT_HOME' -ErrorAction SilentlyContinue }
    if ($EnvKey -eq 'HKCU:\Environment') { Send-SettingChange }
    Ok 'RuneScript is uninstalled. Open a new terminal to update PATH.'
    return
}

# ---- Go ----

Step 'Checking Go'

function Find-Go {
    if ($env:RUNESCRIPT_GO) {
        if (-not (Test-Path -LiteralPath $env:RUNESCRIPT_GO -PathType Leaf)) {
            Fail "RUNESCRIPT_GO is set to '$env:RUNESCRIPT_GO', which is not a file."
        }
        return $env:RUNESCRIPT_GO
    }
    $cmd = Get-Command go -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($cmd) { return $cmd.Source }
    $candidates = @(
        (Join-Path $env:ProgramFiles 'Go\bin\go.exe'),
        'C:\Go\bin\go.exe',
        (Join-Path $env:USERPROFILE 'scoop\shims\go.exe'),
        (Join-Path $env:LOCALAPPDATA 'Programs\Go\bin\go.exe')
    )
    if (${env:ProgramFiles(x86)}) { $candidates += (Join-Path ${env:ProgramFiles(x86)} 'Go\bin\go.exe') }
    foreach ($c in $candidates) {
        if (Test-Path -LiteralPath $c -PathType Leaf) {
            Warn "using $c, which is not on PATH"
            return $c
        }
    }
    Fail @"
Go was not found.
RuneScript is built from source, so the installer needs Go $MinGoText or newer.
Install it from https://go.dev/dl/ (or: winget install GoLang.Go), open a new
terminal, and run the installer again. If Go is installed in an unusual place,
set RUNESCRIPT_GO to the full path of go.exe.
"@
}

function Invoke-Go([string[]]$GoArgs) {
    # Runs go and returns its output (stdout and stderr) and exit code.
    # Windows PowerShell turns native stderr into error records, which
    # would stop the script under ErrorActionPreference = Stop, so it is
    # relaxed here and the records are turned back into text.
    $saved = $env:GOTOOLCHAIN
    $env:GOTOOLCHAIN = 'local'
    $ErrorActionPreference = 'Continue'
    try {
        $out = & $script:Go @GoArgs 2>&1 | ForEach-Object { "$_" }
        return @{ Out = (@($out) -join "`n").Trim(); Code = $LASTEXITCODE }
    } finally {
        $env:GOTOOLCHAIN = $saved
    }
}

$script:Go = Find-Go
$gv = Invoke-Go @('version')
if ($gv.Code -ne 0) {
    Fail "'$script:Go version' failed:`n$($gv.Out)`nThe Go installation looks broken. Reinstall Go from https://go.dev/dl/."
}
# "go version go1.25.11 windows/amd64" (or "go version devel go1.26-abc ...")
if ($gv.Out -notmatch 'go(\d+)\.(\d+)(?:\.(\d+))?') { Fail "could not read the Go version from '$($gv.Out)'" }
$patch = if ($Matches[3]) { $Matches[3] } else { '0' }
$goVersion = [version]"$($Matches[1]).$($Matches[2]).$patch"

# Inside a checkout, go.mod says which Go it needs.
$src = $null
$scriptDir = if ($PSScriptRoot) { $PSScriptRoot } else { $null }
if (-not $Remote -and $scriptDir) {
    $gomod = Join-Path $scriptDir 'go.mod'
    if ((Test-Path -LiteralPath $gomod) -and (Select-String -LiteralPath $gomod -Pattern "^module $([regex]::Escape($Module))$" -Quiet)) {
        $src = $scriptDir
        $need = Select-String -LiteralPath $gomod -Pattern '^go (\d+)\.(\d+)(?:\.(\d+))?' | Select-Object -First 1
        if ($need) {
            $g = $need.Matches[0].Groups
            $MinGo = [version]"$($g[1].Value).$($g[2].Value).$(if ($g[3].Success) { $g[3].Value } else { '0' })"
            $MinGoText = ($need.Line -replace '^go\s+', '').Trim()
        }
    }
}

if ($goVersion -lt $MinGo) {
    Fail "Go $goVersion is too old: RuneScript needs Go $MinGoText or newer.`nUpdate Go from https://go.dev/dl/ and run the installer again."
}
$goroot = Invoke-Go @('env', 'GOROOT')
if ($goroot.Code -ne 0 -or -not $goroot.Out -or -not (Test-Path -LiteralPath $goroot.Out.Trim())) {
    Fail "Go's GOROOT ($($goroot.Out)) is missing; the Go installation looks broken. Reinstall Go from https://go.dev/dl/."
}
Ok "Go $goVersion ($script:Go)"

# ---- build ----

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
# A pure-Go build with this toolchain: no C compiler needed, and no
# toolchain download that would bypass the version check above.
$savedCgo = $env:CGO_ENABLED
$env:CGO_ENABLED = '0'
try {
    if ($src) {
        Step "Building RuneScript from $src"
        Push-Location -LiteralPath $src
        try { $b = Invoke-Go @('build', '-trimpath', '-o', $Target, './cmd/runescript') } finally { Pop-Location }
        if ($b.Code -ne 0) { Fail "the build failed:`n$($b.Out)" }
    } else {
        if (-not (Get-Command git -ErrorAction SilentlyContinue)) { Warn "git is not on PATH; 'go install' may not be able to fetch a branch or commit" }
        Step "Building RuneScript $Version from $Module"
        $tmp = Join-Path ([IO.Path]::GetTempPath()) ("runescript-" + [guid]::NewGuid())
        New-Item -ItemType Directory -Path $tmp | Out-Null
        $savedBin = $env:GOBIN
        $env:GOBIN = $tmp
        try { $b = Invoke-Go @('install', '-trimpath', "$Module/cmd/runescript@$Version") } finally { $env:GOBIN = $savedBin }
        if ($b.Code -ne 0) {
            Remove-Item -LiteralPath $tmp -Recurse -Force
            Fail "'go install $Module/cmd/runescript@$Version' failed:`n$($b.Out)`nCheck your network connection, or that '$Version' is a tag, branch, or commit."
        }
        Move-Item -LiteralPath (Join-Path $tmp 'runescript.exe') -Destination $Target -Force
        Remove-Item -LiteralPath $tmp -Recurse -Force
    }
} finally {
    $env:CGO_ENABLED = $savedCgo
}
Ok "installed $Target"

# ---- check ----

Step 'Checking the installation'
$installed = & $Target version 2>&1
if ($LASTEXITCODE -ne 0) { Fail "the installed CLI does not run: $installed" }
$smoke = Join-Path ([IO.Path]::GetTempPath()) ("runescript-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $smoke | Out-Null
try {
    $hello = "package main;`n`npublic int main() {`n    trace(`"Hello from RuneScript`");`n    return 0;`n}`n"
    [IO.File]::WriteAllText((Join-Path $smoke 'hello.rsc'), $hello, (New-Object Text.UTF8Encoding $false))
    Push-Location -LiteralPath $smoke
    try { $out = & $Target run hello.rsc 2>&1 } finally { Pop-Location }
    if ($LASTEXITCODE -ne 0 -or "$out" -ne 'Hello from RuneScript') { Fail "the installed CLI could not run a test program: $out" }
} finally {
    Remove-Item -LiteralPath $smoke -Recurse -Force
}
Ok "$installed"

if (Get-Command node -ErrorAction SilentlyContinue) {
    Ok "Node.js found: modules from 'runescript build' can run with node"
} else {
    Write-Host "  Node.js not found. It's needed only to run JavaScript from 'runescript build'."
}

# ---- PATH ----

if (-not $NoModifyPath) {
    Step "Adding $BinDir to your user PATH"
    $raw = Get-RawPath
    $parts = @($raw -split ';' | Where-Object { $_ })
    if ($parts | Where-Object { Same-Dir $_ $BinDir }) {
        Ok 'already on PATH'
    } else {
        Set-RawPath ((@($parts) + $BinDir) -join ';')
        Ok 'added'
    }
    if (-not (Test-Path -LiteralPath $EnvKey)) { New-Item -Path $EnvKey -Force | Out-Null }
    Set-ItemProperty -LiteralPath $EnvKey -Name 'RUNESCRIPT_HOME' -Value $InstallDir -Type String
    if ($EnvKey -eq 'HKCU:\Environment') {
        Send-SettingChange
        # This session too.
        if (-not (($env:Path -split ';') | Where-Object { $_ -and (Same-Dir $_ $BinDir) })) { $env:Path = "$env:Path;$BinDir" }
        $env:RUNESCRIPT_HOME = $InstallDir
    }
}

$other = Get-Command runescript -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
if ($other -and -not (Same-Dir (Split-Path $other.Source) $BinDir)) {
    Warn "another runescript is earlier on PATH: $($other.Source)"
}

Write-Host ''
Write-Host 'RuneScript is installed.' -ForegroundColor Green
if ($NoModifyPath) {
    Write-Host "Add $BinDir to PATH to run 'runescript' from anywhere."
} else {
    Write-Host 'Run: runescript version   (in a new terminal, if this one does not find it)'
}
