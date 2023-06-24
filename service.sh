systemctl stop edge-turn.service
echo "[Unit]
Description=
After=network.target

[Service]
Type=simple
ExecStart=/snap/bin/go run cmd/server/main.go
WorkingDirectory=/home/edge-turn

[Install]
WantedBy=multi-user.target" > /lib/systemd/system/edge-turn.service

systemctl enable edge-turn.service
systemctl start  edge-turn.service
systemctl status edge-turn.service