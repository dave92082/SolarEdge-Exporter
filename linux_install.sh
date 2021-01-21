#!/bin/bash
set -e

if [ "$EUID" -ne 0 ]
  then echo "Please run as root!"
  exit
fi

if [ -f "SolarEdge-Exporter" ]; then
    echo "Exporter binary found. Proceeding to install."
else 
    echo "Exporter binary not found! Please copy it here and check if it is named 'SolarEdge-Exporter'"
    exit 1
fi

groupadd --system solaredge_exporter
useradd -s /sbin/nologin --system -g solaredge_exporter solaredge_exporter

cp contrib/solaredge-exporter.service /etc/systemd/system/solaredge_exporter.service
chown root:root /etc/systemd/system/solaredge_exporter.service
chmod 644 /etc/systemd/system/solaredge_exporter.service

cp SolarEdge-Exporter /usr/local/bin/solaredge_exporter
chown solaredge_exporter:solaredge_exporter /usr/local/bin/solaredge_exporter
chmod 755 /usr/local/bin/solaredge_exporter

mkdir /etc/solaredge-exporter
cp config.yaml /etc/solaredge-exporter/config.yaml
chown -R solaredge_exporter:solaredge_exporter /etc/solaredge-exporter

sed -i 's/SolarEdge-Exporter.log/\/var\/log\/SolarEdge\/SolarEdge-Exporter.log/g' /etc/solaredge-exporter/config.yaml

mkdir /var/log/SolarEdge
chown solaredge_exporter:solaredge_exporter /var/log/SolarEdge

echo "All done! Edit config in '/etc/solaredge-exporter/config.yaml' and use '0systemctl enable solaredge_exporter.service' and 'systemctl start solaredge_exporter.service' to start exporter"