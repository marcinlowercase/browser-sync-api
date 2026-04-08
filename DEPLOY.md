GOOS=linux GOARCH=amd64 go build -o browser-sync-api-linux cmd/api-server/main.go

scp -i /path/to/key.pem browser-sync-api-linux schema.sql .env <VPS_USERNAME@<VPS_IP>:~

ssh -i /path/to/key.pem <VPS_USERNAME@<VPS_IP>

pkill -f browser-sync-api-linux

cat schema.sql | sudo -u postgres psql -d browser_sync

chmod +x ~/browser-sync-api-linux

nohup ./browser-sync-api-linux > server.log 2>&1 &

tail -f server.log
