package main

import (
	"context"
	logpkg "log"
	"net/http"
	"os"

	https "github.com/LassiHeikkila/SIM7000/https_native"
	"github.com/LassiHeikkila/mokki-monitoring/mokkimonitoring"
)

func getHTTPClient(ctx context.Context, c mokkimonitoring.Config) *http.Client {
	if c.Comms.UseDefaultClient {
		return http.DefaultClient
	}
	if c.Comms.UseSIM7000 {
		settings := https.Settings{
			APN:        c.Comms.SIM7000Config.APN,
			Username:   c.Comms.SIM7000Config.Username,
			Password:   c.Comms.SIM7000Config.Password,
			SerialPort: c.Comms.SIM7000Config.SerialDevice,
			CertPath:   c.Comms.SIM7000Config.CertificatePath,
		}
		if c.Comms.SIM7000Config.TraceLoggingFile != "" {
			f, err := os.Create(c.Comms.SIM7000Config.TraceLoggingFile)
			if err != nil {
				log.Println("Failed to create file for trace logging modem comms")
			} else {
				tracelogger := logpkg.New(f, "MODEM COMMS: ", logpkg.LstdFlags|logpkg.Lmicroseconds)
				settings.TraceLogger = tracelogger
			}
		}
		httpsclient := https.NewClient(ctx, settings)
		if httpsclient == nil {
			log.Println("Failed to create SIM7000 http transport!")
			return nil
		}
		return &http.Client{
			Transport: httpsclient,
		}
	}
	// nothing configured, use default
	return http.DefaultClient
}
