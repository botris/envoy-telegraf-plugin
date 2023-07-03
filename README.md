# Envoy Telegraf Input Plugin

Gather data from a 3-phase [Enphase Envoy](https://enphase.com/store/communication/envoy-s-metered). <br />
Based on [Telegraf Execd Go Shim](https://github.com/influxdata/telegraf/tree/master/plugins/common/shim).

### Configuration

```toml
[[inputs.envoy]]
    # Envoy management url, replace with IP address if you have assigned your envoy a static IP.
    envoy_url = "http://envoy.local"
```

### Metrics

- envoy
    - fields:
        - total (int)
        - p1_production (int)
        - p1_consumption (int)
        - p1_net (int)
        - p2_production (int)
        - p2_consumption (int)
        - p2_net (int)
        - p3_production (int)
        - p3_consumption (int)
        - p3_net (int)

### Example Output

```
envoy total=939i,p1_consumption=18i,p1_net=-699i,p2_net=-707i,p1_production=718i,p2_production=722i,p2_consumption=14i,p3_production=718i,p3_consumption=3065i,p3_net=2346i 1688391414688034814
envoy p2_production=712i,p3_production=707i,p1_consumption=18i,p1_net=-689i,p2_consumption=15i,p2_net=-696i,p3_consumption=3059i,p3_net=2352i,total=966i,p1_production=708i 1688391415687653261
```

### Build local 
```shell
git clone git@github.com:botris/envoy-telegraf-plugin.git
go build -o build/envoy cmd/main.go
./build/envoy -config plugin.conf 
```
