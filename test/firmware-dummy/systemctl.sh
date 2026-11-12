#!/bin/sh
set -eu
if [ "$#" -eq 2 ] && [ "$1" = "restart" ] && [ "$2" = "griffin.printer" ]; then
    date -u +%Y-%m-%dT%H:%M:%SZ >> /var/log/griffin-restarts.log
    exit 0
fi
echo "dummy systemctl: nicht unterstützter Aufruf: $*" >&2
exit 1
