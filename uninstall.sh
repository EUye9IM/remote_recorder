#!/bin/sh
SRV=/etc/systemd/system/webrtc.service
systemctl stop webrtc
systemctl disable webrtc
rm $SRV
rm /usr/bin/webrtc-server
rm -rf /var/webrtc
rm /etc/webrtc-conf.toml