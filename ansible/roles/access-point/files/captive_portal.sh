#!/bin/bash

service wpa_supplicant stop
sudo killall wpa_supplicant
ip addr add 192.168.4.1/24 dev wlan0
service dnsmasq start
service nftables start
service hostapd restart
