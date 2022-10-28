package hci

import hci "github.com/hypertec-cloud/go-hci"

// Config is the configuration structure used to instantiate a
// new hci client.
type Config struct {
	APIURL   string
	APIKey   string
	Insecure bool
}

// NewClient returns a new HciClient client.
func (c *Config) NewClient() (*hci.HciClient, error) {
	if c.Insecure {
		return hci.NewInsecureHciClientWithURL(c.APIURL, c.APIKey), nil
	}
	return hci.NewHciClientWithURL(c.APIURL, c.APIKey), nil
}
