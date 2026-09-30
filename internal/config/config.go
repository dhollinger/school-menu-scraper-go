package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

var config Config

type Config struct {
	MenuAPIURL string `mapstructure:"menu_api_url"`
	Schools    []struct {
		Name     string `mapstructure:"name"`
		SchoolID string `mapstructure:"school_id"`
	} `mapstructure:"schools"`
	ServingLine   string `mapstructure:"serving_line"`
	MealType      string `mapstructure:"meal_type"`
	Grade         string `mapstructure:"grade"`
	PersonID      string `mapstructure:"person_id"`
	TextbeltPhone string `mapstructure:"textbelt_phone"`
	TextbeltKey   string `mapstructure:"textbelt_key"`
	Date          string `mapstructure:"date"`
}

func Init(path *string) {
	var err error

	v := viper.New()

	if path != nil && *path != "" {
		v.SetConfigFile(*path)
	} else {
		v.SetConfigType("yml")
		v.SetConfigName("menufy")
		v.AddConfigPath(".")
		v.AddConfigPath(".config/menufy")
	}

	err = v.ReadInConfig()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	v = setDefaults(v)

	err = v.Unmarshal(&config)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
}

func setDefaults(v *viper.Viper) *viper.Viper {
	v.SetDefault("menu_api_url", mustEnv("MENU_API_URL"))
	v.SetDefault("serving_line", "Specials of the Day")
	v.SetDefault("meal_type", "Lunch")
	v.SetDefault("grade", mustEnv("GRADE"))
	v.SetDefault("person_id", "null")
	v.SetDefault("textbelt_phone", mustEnv("TEXTBELT_PHONE"))
	v.SetDefault("textbelt_key", mustEnv("TEXTBELT"))
	v.SetDefault("date", time.Now().Format("01/02/2006"))

	return v
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return value
}

func GetConfig() Config {
	return config
}
