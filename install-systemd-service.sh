#!/bin/bash

echo "[..] Building go binary"
go build -ldflags "-s -w"

echo "[..] Making binary executable"
sudo chmod +x ./secure-dns

echo "[..] Copying binary over to /usr/local/bin"
sudo cp ./secure-dns /usr/local/bin

echo "[..] Copying over systemd files"
sudo cp ./systemd/* /etc/systemd/system/

echo "[..] Copying over configuration file"
sudo cp ./conf/secure-dns.conf /etc/secure-dns.conf

echo "[..] Enabling service"
sudo systemctl enable secure-dns

echo "[..] Starting service"
sudo service secure-dns start
