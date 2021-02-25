package mokkimonitoring

import (
	"github.com/LassiHeikkila/SIM7000/http"
	"github.com/LassiHeikkila/SIM7000/module"
)

type Comms struct {
	m module.Module
	c *http.HttpClient
}

func NewComms(settings module.Settings) *Comms {
	m := module.NewSIM7000E(settings)
	if m == nil {
		return nil
	}
	c := http.NewClient(m, http.Settings{APN: settings.APN})
	if c == nil {
		return nil
	}
	return &Comms{m: m, c: c}
}

func (c *Comms) Close() {
	c.c.Close()
	c.m.Close()
}
