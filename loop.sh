while true; do
    go run ./cmd/server/main.go 1> stdout.txt 2>  stderr.txt
    sleep 1 # optional delay
done