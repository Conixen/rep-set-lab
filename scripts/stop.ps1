# Windows Only
$ports = @(8080, 5173)
foreach ($port in $ports) {
    $pids = (Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue).OwningProcess
    foreach ($p in $pids) {
        Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
        Write-Host "Killed process $p on port $port"
    }
}
Write-Host "Done"
