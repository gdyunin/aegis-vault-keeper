#!/bin/bash
# Generate a self-signed TLS certificate and private key for local development.
# Not for production use. Requires OpenSSL. Usage: ./generate_certs.sh
# Outputs: ./certs/server-key.pem (private key), ./certs/server.pem (certificate)

set -e

# Directory where the certificates will be stored
CERTS_DIR="./certs"

# Paths to the generated certificate and private key files
CERT_FILE="$CERTS_DIR/server.pem"
KEY_FILE="$CERTS_DIR/server-key.pem"

# Ensure the output directory exists
mkdir -p "$CERTS_DIR"

echo "Generating self-signed TLS certificates..."

# Generate a 2048-bit RSA private key
openssl genrsa -out "$KEY_FILE" 2048

# Create a self-signed X.509 certificate (365 days) with subject and SANs
# Note: DNS:127.0.0.1 is uncommon but preserved here to match original behavior
openssl req -new -x509 -key "$KEY_FILE" -out "$CERT_FILE" -days 365 \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=AegisVaultKeeper/OU=Development/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,DNS:127.0.0.1,IP:127.0.0.1"

# Print a brief summary for the user
echo "[DONE]: Certificates have been created:"
echo "    Certificate: $CERT_FILE"
echo "    Private Key: $KEY_FILE"
echo "  Validity period: 365 days"
echo "    CN: localhost"
echo "    SAN: localhost, 127.0.0.1"
