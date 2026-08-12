param(
    [Parameter(Mandatory = $true)]
    [string]$Url,
    [Parameter(Mandatory = $true)]
    [string]$Index,
    [Parameter(Mandatory = $true)]
    [string]$From,
    [Parameter(Mandatory = $true)]
    [string]$To,
    [string]$Service = "",
    [ValidateRange(1, 10000)]
    [int]$PageSize = 500,
    [Parameter(Mandatory = $true)]
    [string]$OutputDir,
    [string]$Python = "python",
    [switch]$Insecure
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptPath = Join-Path $PSScriptRoot "extract-from-elasticsearch.py"
$arguments = @(
    $scriptPath,
    "--url", $Url,
    "--index", $Index,
    "--from", $From,
    "--to", $To,
    "--page-size", [string]$PageSize,
    "--output-dir", $OutputDir
)
if (-not [string]::IsNullOrWhiteSpace($Service)) {
    $arguments += @("--service", $Service)
}
if ($Insecure) {
    $arguments += "--insecure"
}

& $Python @arguments
if ($LASTEXITCODE -ne 0) {
    throw "Elasticsearch trace extraction failed with exit code $LASTEXITCODE"
}
