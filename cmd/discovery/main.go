package main

import (
	"errors"
	"github.com/jaku01/caddyservicediscovery/internal/scheduler"
	"github.com/spf13/viper"
	"log"
)

func main() {

	caddyAdminUrl, scheduleInterval, err := loadConfiguration()
	if err != nil {
		panic(err)
	}
	log.Printf("Configuration: CaddyAdminUrl=%s, ScheduleInterval=%d", caddyAdminUrl, scheduleInterval)

	err = scheduler.StartScheduleDiscovery(caddyAdminUrl, scheduleInterval)
	if err != nil {
		panic(err)
	}
}

func loadConfiguration() (string, int, error) {
	viper.SetConfigName("configuration")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("CaddyAdminUrl", "http://localhost:2019")
	viper.SetDefault("ScheduleInterval", 5)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return "", -1, err
		}
		log.Println("No configuration file found, using default values")
	} else {
		log.Println("Configuration file loaded successfully")
	}

	return viper.GetString("CaddyAdminUrl"), viper.GetInt("ScheduleInterval"), nil
}
