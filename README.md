A very lightweight app that connects to the air quality box WP6003 and gather the information via BLE.

- Consumes 10MB RAM
- Tiny docker image 10MB

It collects:

- CO2
- HCHO
- TVOC
- Temperature

ESP32 version is on my blog : [https://omarghader.github.io/esp32-airquality-box-wp6003-homeassistant/](https://omarghader.github.io/esp32-airquality-box-wp6003-homeassistant/)

## Requirements

- Airbox WP6003 Bluetooth : [Buy here from aliexpress](https://s.click.aliexpress.com/e/_AMEXyX)
- Airbox WP6003 Bluetooth : [Buy here from amazon](https://shorturl.servebeer.com/n/airquality-wp6003)

## How to build

```sh
make build

# for arm (raspberry pi)
make build-arm64
```

## How to package by docker

```sh
make build compress docker-build

# for arm (raspberry pi)
make build-arm64 compress docker-build-arm64
```

## How to run

```sh
./bin/airquality -addr=XX:XX:XX:XX:XX:XX
```
