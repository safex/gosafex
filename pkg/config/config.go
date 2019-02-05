package config

import (
	"time"

	"github.com/spf13/viper"
)

// Provider defines a set of read-only methods for accessing the application configuration params as defined in one of the config files
type Provider interface {
	ConfigFileUsed() string
	Get(key string) interface{}
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt64(key string) int64
	GetSizeInBytes(key string) uint
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	InConfig(key string) bool
	IsSet(key string) bool
}

// Default is the global default config ptr
var Default Provider

// LoadConfigProvider will return the config provider for given app name
func LoadConfigProvider(appName string) Provider {
	cfgProvider, err := readViperConfig(appName)
	if err != nil {
		panic(err)
	}
	return cfgProvider
}

// LoadDefault will set the defailt config provider
func LoadDefault(appName string) { Default = LoadConfigProvider(appName) }

func readViperConfig(appName string) (*viper.Viper, error) {
	v := viper.New()

	///////////////////////////////////
	// DEFAULT CONFIG:
	///////////////////////////////////

	v.SetDefault("jsonLogs", false)
	v.SetDefault("loglevel", "debug")

	///////////////////////////////////
	// ENVIRONMENT CONFIG:
	///////////////////////////////////

	v.SetEnvPrefix(appName) // Prefix with <APP_NAME>
	// v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	///////////////////////////////////
	// FILE CONFIG:
	///////////////////////////////////

	v.SetConfigName("config") // name of config file (without extension)
	v.SetConfigType("json")

	v.AddConfigPath(".")

	// TODO: config, system, platform independent
	// TODO: config, user, platform independent
	// TODO: config local, platform independent

	///////////////////////////////////
	// PROCESSING:
	///////////////////////////////////

	if err := v.ReadInConfig(); err != nil {
		NewError("could not read config file", err)
	}

	return v, nil
}
