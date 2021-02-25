package mokkimonitoring

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	InfluxDB       InfluxDBConfig `json:"influxDB"`
	UpdateInterval int            `json:"updateIntervalS"`
	APN            string         `json:"APN"`
	SerialDevice   string         `json:"serialDevice"`
}

type InfluxDBConfig struct {
	URL    string `json:"url"`
	Org    string `json:"org"`
	Bucket string `json:"bucket"`
}

func LoadConfig(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	conf := Config{}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return Config{}, err
	}

	return conf, nil
}
