# Samba Active Directory Domain Controller für die Testumgebung.
FROM debian:12-slim

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
        samba samba-ad-provision samba-dsdb-modules samba-vfs-modules \
        winbind libnss-winbind libpam-winbind \
        krb5-user ldb-tools ldap-utils dnsutils \
    && rm -rf /var/lib/apt/lists/*

COPY docker/samba/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# LDAP(S), Kerberos, DNS, SMB
EXPOSE 389 636 88 464 53 445
ENTRYPOINT ["/entrypoint.sh"]
