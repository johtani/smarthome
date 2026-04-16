param(
    [Parameter(Mandatory = $true)]
    [string]$InputPath,
    [Parameter(Mandatory = $true)]
    [string]$OutputCsv
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-AttrValue {
    param(
        [Parameter(Mandatory = $true)]
        [object[]]$Attributes,
        [Parameter(Mandatory = $true)]
        [string]$Key
    )

    $attr = $Attributes | Where-Object { $_.key -eq $Key } | Select-Object -First 1
    if (-not $attr) {
        return ""
    }
    if ($null -ne $attr.value.stringValue) { return [string]$attr.value.stringValue }
    if ($null -ne $attr.value.intValue) { return [string]$attr.value.intValue }
    if ($null -ne $attr.value.doubleValue) { return [string]$attr.value.doubleValue }
    if ($null -ne $attr.value.boolValue) { return [string]$attr.value.boolValue }
    return ""
}

if (-not (Test-Path -LiteralPath $InputPath)) {
    throw "Input file not found: $InputPath"
}

$rows = New-Object System.Collections.Generic.List[object]

Get-Content -LiteralPath $InputPath | ForEach-Object {
    $line = $_.Trim()
    if ([string]::IsNullOrWhiteSpace($line)) {
        return
    }

    $doc = $line | ConvertFrom-Json
    foreach ($resourceSpan in $doc.resourceSpans) {
        $serviceName = Get-AttrValue -Attributes $resourceSpan.resource.attributes -Key "service.name"
        foreach ($scopeSpan in $resourceSpan.scopeSpans) {
            foreach ($span in $scopeSpan.spans) {
                $traceID = [string]$span.traceId
                foreach ($event in $span.events) {
                    if ($event.name -notin @("resolver.decision", "resolver.execution", "resolver.feedback")) {
                        continue
                    }

                    $attrs = $event.attributes
                    $rows.Add([pscustomobject]@{
                        timestamp                = [string]$event.timeUnixNano
                        service_name             = $serviceName
                        trace_id                 = $traceID
                        event_name               = [string]$event.name
                        resolver_schema_version  = Get-AttrValue -Attributes $attrs -Key "resolver.schema_version"
                        resolver_request_id      = Get-AttrValue -Attributes $attrs -Key "resolver.request_id"
                        resolver_channel         = Get-AttrValue -Attributes $attrs -Key "resolver.channel"
                        resolver_input_text_hash = Get-AttrValue -Attributes $attrs -Key "resolver.input_text_hash"
                        resolver_path            = Get-AttrValue -Attributes $attrs -Key "resolver.path"
                        resolver_mode            = Get-AttrValue -Attributes $attrs -Key "resolver.mode"
                        llm_model                = Get-AttrValue -Attributes $attrs -Key "llm.model"
                        resolver_resolved_command = Get-AttrValue -Attributes $attrs -Key "resolver.resolved_command"
                        resolver_resolved_args   = Get-AttrValue -Attributes $attrs -Key "resolver.resolved_args"
                        resolver_execution_status = Get-AttrValue -Attributes $attrs -Key "resolver.execution_status"
                        feedback_label           = Get-AttrValue -Attributes $attrs -Key "feedback.label"
                        feedback_correction      = Get-AttrValue -Attributes $attrs -Key "feedback.correction"
                        resolver_did_you_mean_command = Get-AttrValue -Attributes $attrs -Key "resolver.did_you_mean_command"
                    })
                }
            }
        }
    }
}

$outputDir = Split-Path -Parent $OutputCsv
if (-not [string]::IsNullOrWhiteSpace($outputDir)) {
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
}

$rows | Export-Csv -LiteralPath $OutputCsv -NoTypeInformation -Encoding UTF8
Write-Output "exported rows: $($rows.Count)"
Write-Output "output: $OutputCsv"
