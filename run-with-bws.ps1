if (Get-BwsToken) {
    bws run -- .\smarthome.exe -server -config-dir config
    Clear-BwsToken
}