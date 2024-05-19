# crypto-feed

## Docker
Build
```
docker build -t crypto-feed .
```

Run
```
docker run -d --rm -v ./config/btcusdt_docker.json:/root/config/config.json -p 50051:50051 crypto-feed
```