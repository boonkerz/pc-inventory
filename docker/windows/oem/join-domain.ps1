# Tritt der Samba-AD-Domaene EXAMPLE.LOCAL bei.
# Einmalig in der Windows-VM ausfuehren (Konsole im Web-Viewer :8006 oder per RDP):
#   powershell -ExecutionPolicy Bypass -File C:\OEM\join-domain.ps1
#
# Voraussetzung: DNS zeigt auf den Samba-DC (setzt install.bat automatisch).

$ErrorActionPreference = "Stop"

$Domain   = "EXAMPLE.LOCAL"
$DcIp     = "172.19.0.10"
$AdminUsr = "EXAMPLE\Administrator"
$AdminPwd = "Passw0rd!"

# DNS sicherheitshalber (erneut) auf den DC setzen.
Get-NetAdapter | Where-Object { $_.Status -eq "Up" } |
    Set-DnsClientServerAddress -ServerAddresses $DcIp

Write-Host "Pruefe DNS-Aufloesung der Domaene ..."
Resolve-DnsName -Type SRV "_ldap._tcp.dc._msdcs.$Domain" -Server $DcIp | Out-Null
Write-Host "Domaenencontroller gefunden. Trete '$Domain' bei ..."

$cred = New-Object System.Management.Automation.PSCredential(
    $AdminUsr, (ConvertTo-SecureString $AdminPwd -AsPlainText -Force))

Add-Computer -DomainName $Domain -Credential $cred -Force
Write-Host "Beitritt erfolgreich. Starte in 5 Sekunden neu ..."
Start-Sleep -Seconds 5
Restart-Computer -Force
