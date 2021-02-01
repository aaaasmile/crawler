package conf

import (
	"encoding/json"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DBPath             string
	DebugSQL           bool
	ServerURI          string
	ServiceAccountFile string
	ServiceAccount     *ServiceAccount
}

type ServiceAccount struct {
	Type             string `json:"type"`
	ProjectID        string `json:"project_id"`
	PrivateKeyID     string `json:"private_key_id"`
	PrivateKey       string `json:"private_key"`
	ClientMail       string `json:"client_email"`
	ClientID         string `json:"client_id"`
	AuthURI          string `json:"auth_uri"`
	TokenURI         string `json:"token_uri"`
	AuthProviderCert string `json:"auth_provider_x509_cert_url"`
	ClientCertURL    string `json:"client_x509_cert_url"`
}

var Current = &Config{}

func ReadConfig(configfile string) (*Config, error) {
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := toml.DecodeFile(configfile, &Current); err != nil {
		log.Fatal(err)
	}

	Current.ServiceAccount, err = readServiceAccountJSONFile(Current.ServiceAccountFile)
	if err != nil {
		return nil, err
	}
	return Current, nil
}

func readServiceAccountJSONFile(cfgFile string) (*ServiceAccount, error) {
	log.Println("Read configuration file for service account ", cfgFile)
	f, err := os.Open(cfgFile)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	info := ServiceAccount{}

	err = json.NewDecoder(f).Decode(&info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}
