package main

import (
	"encoding/hex"
	"flag"
	"io/ioutil"
	logpkg "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"

	"github.com/LassiHeikkila/SIM7000/module"
	"github.com/LassiHeikkila/SIM7000/output"
	"github.com/LassiHeikkila/go-ruuvi/ruuvi"
	"github.com/LassiHeikkila/mokki-monitoring/mokkimonitoring"
)

var ruuviDataHandler func([]byte) = func([]byte) { return }
var log *logpkg.Logger

func init() {
	logpkg.SetOutput(ioutil.Discard)
	log = logpkg.New(os.Stdout, "", logpkg.LstdFlags)
}

func main() {
	log.Println("mokki-monitor started")
	output.SetWriter(log.Writer())
	var (
		configPath = flag.String("conf", "/etc/mokki.json", "Path where config file should be found")
	)
	flag.Parse()

	log.Println("Loading config from:", *configPath)
	conf, err := mokkimonitoring.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config from", *configPath, ":", err)
	}

	token := os.Getenv("INFLUXDB_TOKEN")
	if token == "" {
		log.Fatal("You need to provide INFLUXDB_TOKEN as environment variable")
	}

	moduleSettings := module.Settings{
		APN:                   conf.APN,
		SerialPort:            conf.SerialDevice,
		MaxConnectionAttempts: 15,
	}
	log.Println("opening comms...")
	c := mokkimonitoring.NewComms(moduleSettings)
	if c == nil {
		log.Println("Failed to open comms")
		return
	}
	defer c.Close()

	advertChan := make(chan ruuvi.AdvertisementData)

	ruuviDataHandler = func(b []byte) {
		if len(b) == 0 {
			return
		}
		log.Println("Device has advertisement payload:", hex.EncodeToString(b))

		advert, err := ruuvi.ProcessAdvertisement(b)
		if err != nil {
			log.Println("Error processing bytes:", err)
			return
		}
		advert.Copy()

		advertChan <- advert
	}

	log.Println("starting bluetooth comms...")

	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Printf("Failed to open device, err: %s\n", err)
		os.Exit(1)
	}

	// Register handlers.
	d.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	d.Init(onStateChanged)

	sc := make(chan os.Signal)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sc:
			d.StopScanning()
			return
		case advert := <-advertChan:
			log.Println("Handling advert")
			lp, err := mokkimonitoring.RuuviDataToInfluxDBLineProtocol(advert)
			if err != nil {
				log.Println("Error transforming advert to line protocol:", err)
				continue

			}
			lp_data, err := lp.Marshal()
			if err != nil {
				log.Println("Error marshalling line protocol struct:", err)
				continue
			}
			log.Println("lp:", string(lp_data))
			err = mokkimonitoring.PostToInfluxDB(c, lp_data, conf.InfluxDB.URL, conf.InfluxDB.Org, conf.InfluxDB.Bucket, token)
			if err != nil {
				log.Println("Error posting data to InfluxDB:", err)
				continue
			}
		}
	}
}

func onStateChanged(d gatt.Device, s gatt.State) {
	log.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Println("scanning...")
		d.Scan([]gatt.UUID{}, true) // report duplicates since we want to keep receiving adverts from sensors
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if !ruuvi.IsAdvertisementFromRuuviTag(a.ManufacturerData) {
		return
	}

	ruuviDataHandler(a.ManufacturerData)
}
