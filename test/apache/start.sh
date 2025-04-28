#!/usr/bin/env bash

# If the transient local TLS certificate doesn't exist, mint a new one
if [ ! -f ../pki/httpd.local.cetacean.club/cert.pem ]; then
  # Subshell to contain the directory change
  (
    cd ../pki \
    && mkdir -p httpd.local.cetacean.club \
    && \
    # Try using https://github.com/FiloSottile/mkcert for better DevEx,
    # but fall back to using https://github.com/jsha/minica in case
    # you don't have that installed.
    (
      mkcert \
        --cert-file ./httpd.local.cetacean.club/cert.pem \
        --key-file ./httpd.local.cetacean.club/key.pem httpd.local.cetacean.club \
      || go tool minica -domains httpd.local.cetacean.club
    )
  )
fi

docker compose up --build