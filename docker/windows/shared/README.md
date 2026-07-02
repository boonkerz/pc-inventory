# Shared-Folder für die Windows-VM

Der Inhalt dieses Ordners ist in der Windows-VM als Netzwerkfreigabe **`\\172.30.0.1\Data`**
erreichbar (dockur). Praktisch, um nachträglich Dateien in eine laufende VM zu schieben –
z. B. die `agent.exe`, ohne neu zu installieren.

> Die Freigabe heißt bei dockur eigentlich `\\host.lan\Data`. Da die VM hier den Samba-DC
> als DNS nutzt, wird der Name `host.lan` nicht aufgelöst – daher die IP **`172.30.0.1`**
> (das dockur-interne Gateway, auf dem `smbd` lauscht).

Damit dockur diesen Ordner als Freigabe nutzt, ist er in `compose.yml` als `/data` gemountet.
Greift nach einem (Neu-)Start des `windows-test-pc`-Containers.

> Bei einer laufenden VM ohne diesen Mount geht es auch direkt:
> `docker cp datei pc-inventory-test-windows-test-pc-1:/tmp/smb/` – die Datei erscheint
> sofort unter `\\172.30.0.1\Data`.
