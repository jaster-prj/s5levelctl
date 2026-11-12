#!/bin/sh
set -eu
target=/usr/share/griffin/griffin/printer/procedures/pre_and_post_print/auto_bed_level_adjust/alignZAxisProcedure.py
mkdir -p "$(dirname "$target")"
cp /opt/s5levelctl-dummy/alignZAxisProcedure.py "$target"
rm -f "$target".s5levelctl-backup-* "$target".ultilevel-backup-*
rm -f /var/log/griffin-restarts.log
