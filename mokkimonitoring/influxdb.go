package mokkimonitoring

import (
	"encoding/hex"
	"time"

	"github.com/LassiHeikkila/go-ruuvi/ruuvi"
	writeapi "github.com/influxdata/influxdb-client-go/v2/api/write"
)

func RuuviDataToInfluxDBPoint(ad ruuvi.AdvertisementData) (string, *writeapi.Point, error) {
	point := writeapi.NewPointWithMeasurement("ruuvidata")
	point.SetTime(time.Now().UTC())

	var macaddress string
	if mac, err := ad.MACAddress(); err == nil {
		macaddress = hex.EncodeToString(mac)
		point.AddTag("sensormac", macaddress)
	}
	if temp, err := ad.Temperature(); err == nil {
		point.AddField("temperature", temp)
	}
	if humidity, err := ad.Humidity(); err == nil {
		point.AddField("humidity", humidity)
	}
	if pressure, err := ad.Pressure(); err == nil {
		point.AddField("pressure", pressure)
	}
	if batteryvolts, err := ad.BatteryVoltage(); err == nil {
		point.AddField("batteryvoltage", batteryvolts)
	}

	return macaddress, point, nil
}

/*
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
	req, err := http.NewRequestWithContext(comms.ctx, "POST", postUrl, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header["AUTHORIZATION"] = []string{fmt.Sprintf("Token %s", token)}

	resp, err := comms.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Server replied:", string(body))
		return fmt.Errorf("Not OK status code returned: %d (%s)", resp.StatusCode, resp.Status)
	}

	return nil
}
*/
