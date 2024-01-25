package server

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/viper"
)

func NewLogger() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	level, err := zerolog.ParseLevel(viper.GetString("logger.level"))
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	zerolog.SetGlobalLevel(level)
}
