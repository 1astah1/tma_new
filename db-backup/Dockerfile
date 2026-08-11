FROM postgres:15-alpine

COPY backup.sh /usr/local/bin/backup.sh
RUN chmod +x /usr/local/bin/backup.sh

VOLUME /backups
ENTRYPOINT ["/usr/local/bin/backup.sh"]
