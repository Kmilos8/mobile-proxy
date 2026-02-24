#!/bin/bash
# Initial OpenVPN PKI setup script
# Run once to initialize certificates

set -e

OVPN_DATA="/etc/openvpn"
SERVER_IP="${1:?Usage: setup.sh <server-public-ip>}"

echo "Initializing OpenVPN PKI..."

# Generate PKI
docker run -v openvpn_data:/etc/openvpn --rm kylemanna/openvpn ovpn_genconfig -u udp://$SERVER_IP
docker run -v openvpn_data:/etc/openvpn --rm -it kylemanna/openvpn ovpn_initpki

# Create CCD directory
docker run -v openvpn_data:/etc/openvpn --rm kylemanna/openvpn mkdir -p /etc/openvpn/ccd

echo "OpenVPN PKI initialized!"
echo ""
echo "To generate a client certificate for a device:"
echo "  docker run -v openvpn_data:/etc/openvpn --rm -it kylemanna/openvpn easyrsa build-client-full <device-name> nopass"
echo "  docker run -v openvpn_data:/etc/openvpn --rm kylemanna/openvpn ovpn_getclient <device-name> > <device-name>.ovpn"
