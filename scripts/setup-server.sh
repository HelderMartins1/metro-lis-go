#!/usr/bin/env bash
# Prepares the server for Cais, once. Safe to run again: it never overwrites
# an existing /var/www/cais/.env or the push data.
#
# From your machine, in the repository:
#   scp deploy/cais.service deploy/nginx/caislx.tech.conf .env.example <user>@<vm>:/tmp/
#   ssh <user>@<vm> 'sudo bash -s' < scripts/setup-server.sh
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "run as root (sudo)" >&2
  exit 1
fi

# A user of its own, with no shell, that owns only the push data.
if ! id cais >/dev/null 2>&1; then
  useradd --system --no-create-home --shell /usr/sbin/nologin cais
fi

install -d -m 755 -o root -g root /var/www/cais /var/www/cais/www
install -d -m 700 -o cais -g cais /var/www/cais/data

# .env holds the Metro API token: readable by the service only.
if [ ! -f /var/www/cais/.env ]; then
  install -m 600 -o cais -g cais /tmp/.env.example /var/www/cais/.env
  echo "Created /var/www/cais/.env from the example: fill in METRO_API_URL and METRO_API_TOKEN."
fi

install -m 644 /tmp/cais.service /etc/systemd/system/cais.service
systemctl daemon-reload
systemctl enable cais
echo "Installed the cais service; it starts on the first deploy."

if [ -d /etc/nginx/sites-available ]; then
  install -m 644 /tmp/caislx.tech.conf /etc/nginx/sites-available/caislx.tech
  ln -sf /etc/nginx/sites-available/caislx.tech /etc/nginx/sites-enabled/caislx.tech
  nginx -t && systemctl reload nginx
  echo "nginx serves caislx.tech and api.caislx.tech over http; run certbot next for https."
fi

rm -f /tmp/cais.service /tmp/caislx.tech.conf /tmp/.env.example
echo "Done."
