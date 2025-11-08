#!/bin/bash

# Create certs directory if it doesn't exist
mkdir -p certs

# Generate private key
openssl genrsa -out certs/server.key 2048

# Generate self-signed certificate
openssl req -new -x509 -key certs/server.key -out certs/server.crt -days 365 -subj '/C=JP/ST=Tokyo/L=Tokyo/O=Magic-Cylinder/OU=IT Department/CN=localhost'

echo "Certificates generated successfully in certs/ directory"