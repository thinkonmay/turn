go build -o turn cmd/server/main.go
systemctl stop edge-turn-huyhoangdo.service
echo "[Unit]
Description=
After=network.target

StartLimitIntervalSec=500
StartLimitBurst=5

[Service]
Type=simple
ExecStart=/home/ubuntu/edge-turn-huyhoangdo/turn
WorkingDirectory=/home/ubuntu/edge-turn-huyhoangdo

Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target" > /lib/systemd/system/edge-turn-huyhoangdo.service

systemctl enable edge-turn-huyhoangdo.service
systemctl start  edge-turn-huyhoangdo.service
systemctl status edge-turn-huyhoangdo.service