package mokkimonitoring

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	InfluxDB       InfluxDBConfig `json:"influxDB"`
	UpdateInterval int            `json:"updateIntervalS"`
	Comms          CommsConfig    `json:"comms"`
}

type InfluxDBConfig struct {
	URL    string `json:"url"`
	Org    string `json:"org"`
	Bucket string `json:"bucket"`
	Token  string `json:"token"`
}

type CommsConfig struct {
	UseDefaultClient bool          `json:"useDefaultClient"`
	UseSIM7000       bool          `json:"useSIM7000"`
	SIM7000Config    Sim7000Config `json:"sim7000"`
}

type Sim7000Config struct {
	APN              string `json:"apn"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	SerialDevice     string `json:"serialDevice"`
	CertificatePath  string `json:"certPath"`
	TraceLoggingFile string `json:"traceLoggingFile"`
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
