package main

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	logpkg "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	dbclient "github.com/influxdata/influxdb-client-go/v2"
	httpapi "github.com/influxdata/influxdb-client-go/v2/api/http"
	writeapi "github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"

	"github.com/LassiHeikkila/SIM7000/output"
	"github.com/LassiHeikkila/go-ruuvi/ruuvi"
	"github.com/LassiHeikkila/mokki-monitoring/mokkimonitoring"
)

var ruuviDataHandler = func([]byte) { return }
var log *logpkg.Logger

func init() {
	logpkg.SetOutput(ioutil.Discard)
	log = logpkg.New(os.Stdout, "", logpkg.LstdFlags|logpkg.Llongfile|logpkg.Lmicroseconds)
}

func main() {
	log.Println("mokki-monitor started")
	output.SetWriter(log.Writer())
	var (
		configPath = flag.String("conf", "/etc/mokki.json", "Path where config file should be found")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Loading config from:", *configPath)
	conf, err := mokkimonitoring.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config from", *configPath, ":", err)
	}
	log.Println("opening comms...")
	httpclient := getHTTPClient(ctx, conf)
	if httpclient == nil {
		log.Println("Failed to open comms")
		return
	}

	advertChan := make(chan ruuvi.AdvertisementData)

	ruuviDataHandler = func(b []byte) {
		if len(b) == 0 {
			return
		}
		//log.Println("Device has advertisement payload:", hex.EncodeToString(b))
		advert, err := ruuvi.ProcessAdvertisement(b)
		if err != nil {
			log.Println("Error processing bytes:", err)
			return
		}
		advert.Copy()

		advertChan <- advert
	}

	log.Println("Starting bluetooth comms...")

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
	go func() {
		alreadyCalledOnce := false
		for {
			<-sc
			if !alreadyCalledOnce {
				log.Println("Exit requested, cancelling context!")
				cancel()
				alreadyCalledOnce = true
			} else {
				log.Println("Graceful exit did not work in time, calling os.Exit(1)!")
				os.Exit(1)
			}
		}
	}()

	timeToPost := time.NewTicker(time.Duration(conf.UpdateInterval) * time.Second)

	latestPoints := make(map[string]*writeapi.Point)

	printState := func() {
		for mac, p := range latestPoints {
			log.Printf("Latest data point for MAC %s:\n", mac)
			fields := p.FieldList()
			for _, f := range fields {
				if f == nil {
					continue
				}
				log.Printf("\t%v\n", *f)
			}
		}
	}

	opts := dbclient.DefaultOptions().SetHTTPClient(httpclient).SetHTTPRequestTimeout(30) // timeout given in seconds
	influxdbclient := dbclient.NewClientWithOptions(conf.InfluxDB.URL, conf.InfluxDB.Token, opts)
	defer influxdbclient.Close()

	writeAPI := influxdbclient.WriteAPIBlocking(conf.InfluxDB.Org, conf.InfluxDB.Bucket)

	consequtiveWriteErrors := 0
	writeErrorLimit := 10
eventloop:
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping bluetooth scanning.")
			d.StopScanning()
			break eventloop
		case advert := <-advertChan:
			//log.Println("Handling advert")
			mac, p, err := mokkimonitoring.RuuviDataToInfluxDBPoint(advert)
			if err != nil {
				log.Println("Error transforming advert to line protocol:", err)
				continue
			}
			latestPoints[mac] = p
		case <-timeToPost.C:
			printState()
			log.Println("Writing data to InfluxDB")
			for _, p := range latestPoints {
				//log.Printf("point: %#v\n", p)
				err = writeAPI.WritePoint(ctx, p)
				if err != nil {
					if errors.Is(err, &httpapi.Error{}) {
						// we can tolerate this error, it seems like influxdb gives back 404 every time :/
						continue
					}
					log.Println("Error writing record:", err)
					consequtiveWriteErrors++
					if consequtiveWriteErrors >= writeErrorLimit {
						log.Println("Too many consequtive errors, exiting!")
						return
					}
				} else {
					consequtiveWriteErrors = 0
				}
			}
		}
	}
	log.Println("Exiting")
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
