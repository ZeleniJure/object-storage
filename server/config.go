package server

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Config() {
	defaults()
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal().Err(err).Msg("Configuration file not found 'config.yaml'")
		} else {
			log.Fatal().Err(err).Msg("Could not load configuration file 'config.yaml'")
		}
	}
	viper.SetEnvPrefix("obstore")
	viper.AutomaticEnv()
}

func defaults() {
	// Not setting any defaults, since we are shipping with
	// a config file already.
}
