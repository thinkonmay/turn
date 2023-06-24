go build -o turn cmd/server/main.go
systemctl stop edge-turn.service
echo "[Unit]
Description=
After=network.target

StartLimitIntervalSec=500
StartLimitBurst=5

[Service]
Type=simple
ExecStart=/home/edge-turn/turn
WorkingDirectory=/home/edge-turn

Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target" > /lib/systemd/system/edge-turn.service

systemctl enable edge-turn.service
systemctl start  edge-turn.service
systemctl status edge-turn.service