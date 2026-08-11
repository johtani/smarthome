param(
    [Parameter(Mandatory = $true)]
    [string]$ResolverEventsCsv,
    [Parameter(Mandatory = $false)]
    [string]$WorkDir = ".\\tmp\\dspy",
    [Parameter(Mandatory = $false)]
    [string]$Model = "openai/gpt-4o-mini",
    [Parameter(Mandatory = $false)]
    [string]$PromptVersion = "offline-eval-v1",
    [Parameter(Mandatory = $false)]
    [string]$ApiBase = "",
    [Parameter(Mandatory = $false)]
    [string]$ApiKey = "",
    [Parameter(Mandatory = $false)]
    [string]$ModelType = "",
    [Parameter(Mandatory = $false)]
    [Nullable[double]]$Temperature = $null,
    [Parameter(Mandatory = $false)]
    [Nullable[int]]$MaxTokens = $null
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null

$datasetPath = Join-Path $WorkDir "dataset.jsonl"
$reportPath = Join-Path $WorkDir "report.json"

python tools/dspy/prepare_dataset.py `
  --input-csv $ResolverEventsCsv `
  --output-jsonl $datasetPath `
  --min-row-per-request 2

$optimizeArgs = @(
  "tools/dspy/optimize_and_evaluate.py",
  "--dataset-jsonl", $datasetPath,
  "--command-catalog", "tools/dspy/command_catalog.sample.json",
  "--model", $Model,
  "--prompt-version", $PromptVersion,
  "--report-out", $reportPath,
  "--min-command-accuracy", "0.80",
  "--min-arg-accuracy", "0.60"
)

if ($ApiBase) {
  $optimizeArgs += @("--api-base", $ApiBase)
}
if ($ApiKey) {
  $optimizeArgs += @("--api-key", $ApiKey)
}
if ($ModelType) {
  $optimizeArgs += @("--model-type", $ModelType)
}
if ($null -ne $Temperature) {
  $optimizeArgs += @("--temperature", $Temperature.ToString([System.Globalization.CultureInfo]::InvariantCulture))
}
if ($null -ne $MaxTokens) {
  $optimizeArgs += @("--max-tokens", $MaxTokens.ToString([System.Globalization.CultureInfo]::InvariantCulture))
}

python @optimizeArgs

Write-Output "batch finished: $reportPath"
