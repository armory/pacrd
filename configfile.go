package main

import (
	"github.com/armory/go-yaml-tools/pkg/tls/client"
	"github.com/armory/go-yaml-tools/pkg/tls/server"
	"github.com/spf13/viper"
)

// PacrdConfig represents the valid configuration options for PaCRD
type PacrdConfig struct {
	// TrimSpaceTlsPassword (optional) used when parsing tls password to delete spaces and new lines
	TrimSpaceTlsPassword bool
	// NewRelicLicense (optional) license.
	NewRelicLicense string
	// NewRelicApp name (optional)
	NewRelicAppName string
	// FiatServiceAccount (optional) the service account to annotate API calls with.
	FiatServiceAccount string
	// SpinnakerServices defines the location of Spinnaker services in your environment.
	SpinnakerServices
	//Server config for mTLS
	Server server.ServerConfig `json:"server" yaml:"server"`
	Http   client.Config       `json:"http" yaml:"http"`
}

// SpinnakerServices represent the set of services that PaCRD must interface with.
type SpinnakerServices struct {
	Front50 string
	Orca    string
}

// InitConfig initializes configuration for PaCRD.
func InitConfig() (PacrdConfig, error) {

	var config PacrdConfig

	viper.SetConfigName("pacrd")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/opt/pacrd")
	viper.SetDefault("NewRelicAppName", "pacrd")
	viper.SetDefault("TrimSpaceTlsPassword", true)
	viper.SetDefault("SpinnakerServices", SpinnakerServices{
		Front50: "http://spin-front50:8080",
		Orca:    "http://spin-orca:8083",
	})
	viper.SetDefault("FiatServiceAccount", "")

	if err := viper.ReadInConfig(); err != nil {
		// Ignore not found errors; we'll use the defaults in this case.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return PacrdConfig{}, err
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return PacrdConfig{}, err
	}

	return config, nil
}
