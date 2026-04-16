param(
    [Parameter(Mandatory = $true)]
    [string]$ResolverEventsCsv,
    [Parameter(Mandatory = $false)]
    [string]$WorkDir = ".\\tmp\\dspy",
    [Parameter(Mandatory = $false)]
    [string]$Model = "openai/gpt-4o-mini"
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

python tools/dspy/optimize_and_evaluate.py `
  --dataset-jsonl $datasetPath `
  --command-catalog tools/dspy/command_catalog.sample.json `
  --model $Model `
  --report-out $reportPath `
  --min-command-accuracy 0.80 `
  --min-arg-accuracy 0.60

Write-Output "batch finished: $reportPath"
