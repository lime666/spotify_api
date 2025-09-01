package config

import "github.com/spf13/viper"

type Config struct {
	SpotifyClientID     string `mapstructure:"SPOTIFY_CLIENT_ID"`
	SpotifyClientSecret string `mapstructure:"SPOTIFY_CLIENT_SECRET"`
	SpotifyRedirectURL  string `mapstructure:"SPOTIFY_REDIRECT_URL"`
	Port                string `mapstructure:"PORT"`
}

func LoadConfig() (c Config, err error) {
	viper.AddConfigPath("./config/envs")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	err = viper.Unmarshal(&c)
	return
}
