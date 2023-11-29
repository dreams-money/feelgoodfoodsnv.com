package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Environment        string   `json:"environment"`
	HttpPort           int      `json:"http_port"`
	TLSPort            int      `json:"tls_port"`
	Hosts              []string `json:"hosts"`
	GoogleMapsAPIKeys  EnvKeys  `json:"google_maps_api_key"`
	PublicPaymentKeys  EnvKeys  `json:"payments_public_key"`
	PrivatePaymentKeys EnvKeys  `json:"payments_private_key"`
	MailerKeys         EnvKeys  `json:"mailer_key"`
	MailerFromName     EnvKeys  `json:"mailer_from_name"`
	MailerFromEmail    EnvKeys  `json:"mailer_from_email"`
	AcceptOrderDays    []int    `json:"accept_order_days"`
}

type EnvKeys struct {
	Production  string `json:"prd"`
	Development string `json:"dev"`
}

func LoadConfig(fileName string) Config {
	var conf Config

	file, err := os.Open(fileName)
	panicOnError(err)

	fileBytes, err := ioutil.ReadAll(file)
	panicOnError(err)

	json.Unmarshal(fileBytes, &conf)

	return conf
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
