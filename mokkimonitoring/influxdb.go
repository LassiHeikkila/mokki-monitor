package mokkimonitoring

import (
	"encoding/hex"
	"fmt"
	"log"
	urlpkg "net/url"
	"time"

	"github.com/LassiHeikkila/mokki-monitoring/influxdb"

	"github.com/LassiHeikkila/go-ruuvi/ruuvi"
)

func RuuviDataToInfluxDBLineProtocol(ad ruuvi.AdvertisementData) (influxdb.LineProtocol, error) {
	lp := influxdb.LineProtocol{
		Measurement: "ruuvidata",
		TagSet:      make(map[string]string),
		FieldSet:    make(map[string]influxdb.FieldValue),
		Timestamp:   time.Now().UTC(),
	}
	if mac, err := ad.MACAddress(); err == nil {
		lp.TagSet["sensormac"] = hex.EncodeToString(mac)
	}
	if temp, err := ad.Temperature(); err == nil {
		lp.FieldSet["temperature"] = influxdb.NewFieldValue(temp)
	}
	if humidity, err := ad.Humidity(); err == nil {
		lp.FieldSet["humidity"] = influxdb.NewFieldValue(humidity)
	}
	if pressure, err := ad.Pressure(); err == nil {
		lp.FieldSet["pressure"] = influxdb.NewFieldValue(pressure)
	}
	if batteryvolts, err := ad.BatteryVoltage(); err == nil {
		lp.FieldSet["batteryvoltage"] = influxdb.NewFieldValue(batteryvolts)
	}

	return lp, nil
}

func PostToInfluxDB(
	comms *Comms,
	data []byte,
	url string,
	org string,
	bucket string,
	token string,
) error {
	vals := urlpkg.Values{}
	vals.Set("org", org)
	vals.Set("bucket", bucket)
	postUrl := fmt.Sprintf("%s/%s", url, vals.Encode())
	status, reply, err := comms.c.Post(postUrl, data, map[string]string{"AUTHORIZATION": fmt.Sprintf("Token %s", token)})
	if err != nil {
		return err
	}
	if status != 200 {
		log.Println("Server replied:", reply)
		return fmt.Errorf("Not OK status code returned: %d", status)
	}

	return nil
}
