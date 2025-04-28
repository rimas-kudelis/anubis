#!/usr/bin/env bash

# If the transient local TLS certificate doesn't exist, mint a new one
if [ ! -f ../pki/nginx.local.cetacean.club/cert.pem ]; then
  # Subshell to contain the directory change
  (
    cd ../pki \
    && mkdir -p nginx.local.cetacean.club \
    && \
    # Try using https://github.com/FiloSottile/mkcert for better DevEx,
    # but fall back to using https://github.com/jsha/minica in case
    # you don't have that installed.
    (
      mkcert \
        --cert-file ./nginx.local.cetacean.club/cert.pem \
        --key-file ./nginx.local.cetacean.club/key.pem nginx.local.cetacean.club \
      || go tool minica -domains nginx.local.cetacean.club
    )
  )
fi

docker compose up --build