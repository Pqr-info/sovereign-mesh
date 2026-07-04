# PowerShell Bridge for WSL -> Windows (gRPC and High-Speed Bus)
# -------------------------------------------------------------
# Configures Windows port-proxy rules to forward ports 1111 (gRPC) and 11111 (Neural Gossip Bus)
# to the corresponding services running inside WSL2. Also optionally starts the PQR server.

# ---- Configuration ----
$GrpcPort = 1111   # gRPC Control Plane
$BusPort  = 11111  # Neural Gossip Bus

# Get the correct IPv4 address of the WSL2 distro
$WslIp = ((wsl hostname -I).Trim() -split '\s+') | Where-Object { 
    $_ -match '^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$' -and 
    $_ -ne '172.17.0.1' -and 
    $_ -ne '127.0.0.1' 
} | Select-Object -First 1

Write-Host "[Bridge] Detected WSL IP: $WslIp" -ForegroundColor Cyan

function Add-Proxy {
    param([int]$listenPort, [int]$connectPort)
    $cmd = "netsh interface portproxy add v4tov4 listenaddress=0.0.0.0 listenport=$listenPort connectaddress=${WslIp} connectport=$connectPort"
    Write-Host "[Bridge] Adding proxy: $listenPort -> ${WslIp}:$connectPort" -ForegroundColor Green
    Invoke-Expression $cmd
}

function Remove-Existing {
    param([int]$port)
    $cmd = "netsh interface portproxy delete v4tov4 listenport=$port listenaddress=0.0.0.0"
    Invoke-Expression $cmd 2>$null
}

# Clean old rules
Remove-Existing $GrpcPort
Remove-Existing $BusPort

# Add new forwarding rules
Add-Proxy $GrpcPort $GrpcPort
Add-Proxy $BusPort $BusPort

Write-Host "[Bridge] Current portproxy list:" -ForegroundColor Yellow
netsh interface portproxy show all

function Test-ConnectionPort {
    param([int]$port)
    try {
        $tcp = New-Object Net.Sockets.TcpClient "localhost", $port
        $tcp.Close()
        Write-Host "[Bridge] Port $port is reachable (proxy active)." -ForegroundColor Green
    } catch {
        Write-Host "[Bridge] Port $port not reachable - check WSL service." -ForegroundColor Red
    }
}

Test-ConnectionPort $GrpcPort
Test-ConnectionPort $BusPort

$launch = Read-Host "Do you want to start the PQR server inside WSL now? (y/N)"
if ($launch -eq 'y' -or $launch -eq 'Y') {
    wsl -e bash -c 'cd /home/aellok/pqr-info-swarm ; ./pqr-server'
    Write-Host "[Bridge] PQR server started in WSL." -ForegroundColor Magenta
}

Write-Host "[Bridge] Setup complete. Windows can now reach gRPC on port $GrpcPort and the bus on $BusPort via localhost." -ForegroundColor Cyan
