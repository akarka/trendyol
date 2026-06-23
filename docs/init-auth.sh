#!/bin/sh
# docker-entrypoint-initdb.d script: app kullanıcısının auth plugin'ini
# mysql_native_password'e çevirir. MySQL 8 varsayılanı caching_sha2_password;
# internal/server/admin.go'nun çalıştırdığı mysqldump/mysql CLI (alpine
# mariadb-client) bu plugin'in shared library'sini içermiyor. Go driver
# (pure Go, internal/db/db.go) caching_sha2_password'ü zaten destekliyor —
# bu değişiklik sadece CLI ile yedekleme/geri yükleme içindir.
set -e

mysql -uroot -p"$MYSQL_ROOT_PASSWORD" <<-EOSQL
  ALTER USER IF EXISTS '$MYSQL_USER'@'%' IDENTIFIED WITH mysql_native_password BY '$MYSQL_PASSWORD';
EOSQL
