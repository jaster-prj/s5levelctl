FROM alpine:3.22

RUN apk add --no-cache openssh-server python3 \
    && echo 'root:ultimaker' | chpasswd \
    && mkdir -p /run/sshd

COPY sshd_config /etc/ssh/sshd_config
COPY alignZAxisProcedure.py /opt/s5levelctl-dummy/alignZAxisProcedure.py
COPY ultimaker_firmware /etc/ultimaker_firmware
COPY entrypoint.sh /usr/local/bin/dummy-entrypoint
COPY systemctl.sh /usr/local/bin/systemctl
COPY reset-dummy.sh /usr/local/bin/reset-dummy

RUN chmod +x /usr/local/bin/dummy-entrypoint /usr/local/bin/systemctl /usr/local/bin/reset-dummy \
    && /usr/local/bin/reset-dummy

EXPOSE 22
HEALTHCHECK --interval=2s --timeout=2s --start-period=2s --retries=15 \
    CMD python3 -c "import socket;s=socket.create_connection(('127.0.0.1',22),2);s.close()"

ENTRYPOINT ["/usr/local/bin/dummy-entrypoint"]
