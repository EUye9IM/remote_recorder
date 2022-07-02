#!/bin/sh
SRV=/etc/systemd/system/webrtc.service

cd ./source

# toolchain
dnf install golang -y

# compile
bash build.sh

# install
mkdir -p /usr/bin
cp ./webrtc-server /usr/bin
mkdir -p /var/webrtc
cp -r ./sign /var/webrtc/sign
cp -r ./resources /var/webrtc/resources
cp ../config/webrtc-conf.toml /etc/webrtc-conf.toml


echo "[Unit]" > $SRV
echo "Description=just big homework about webrtc" >> $SRV
echo "[Service]" >> $SRV
echo "Type=simple" >> $SRV
echo "ExecStart=webrtc-server -c /etc/webrtc-conf.toml" >> $SRV
echo "[Install]" >> $SRV
echo "WantedBy=multi-user.target" >> $SRV

# set service
systemctl daemon-reload
systemctl start webrtc
systemctl enable webrtc