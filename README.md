# mokki-monitoring
Mökki is a finnish term for a cabin. It is often left uninhabited for long periods of time in winter.

This application combines wireless environmental sensors with low power usage,
low cost internet connectivity via a NB-IoT / GPRS module and a low cost computer to provide DIY
environment monitoring at a remote cabin, visible through an InfluxDB dashboard.

## Sensors
Supported sensors are Ruuvitags, which are little discs sending out BLE advertisements containing raw data, and Ruuvi Air, which is an indoor air quality measurement device.
This is handled with github.com/LassiHeikkila/go-ruuvi library.
BLE packets are picked up by the Raspberry Pi bluetooth device and parsed with the library.

## Internet connectivity
Internet connectivity is handled by a Waveshare SIM7000E hat for a Raspberry Pi 4 (or other),
wrapped by github.com/LassiHeikkila/SIM7000
The library currently supports HTTP GET and POST, which should be enough for this application.
System default network connection can also be used.

## InfluxDB
Parsed data is POSTed to influxdb. Parameters like URL, org name, bucket and auth token are given with config file (see Config section).

## Config
Config is JSON formatted file, default path is `/etc/mokki.json`, but path can be changed with `-conf` flag
Example contents:
```json
{
	"influxDB": {
		"url": "http://example.com:8099/",
		"org": "org@example.com",
		"bucket": "data-bucket",
		"token": "my secret API token"
	},
	"updateIntervalS": 30,
	"comms": {
		"useDefaultClient": false,
		"useSIM7000": true,
		"sim7000": {
			"apn": "internet",
			"serialDevice": "/dev/ttyS0",
			"traceLoggingFile": "/var/log/sim7000trace.log"
		}
	}
}
```

Another example using system default HTTP communication:
```json
{
	"influxDB": {
		"url": "http://example.com:8099/",
		"org": "org@example.com",
		"bucket": "data-bucket",
		"token": "my secret API token"
	},
	"updateIntervalS": 30,
	"comms": {
		"useDefaultClient": true,
		"useSIM7000": false,
		"sim7000": {}
	}
}
```
