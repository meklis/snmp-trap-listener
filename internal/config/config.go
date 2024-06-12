package config

import (
	"embed"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

//go:embed   trap-listener.yml
var configFile embed.FS

// LoggerConfig represents the configuration for the logger.
type LoggerConfig struct {
	Console struct {
		Enabled     bool   `yaml:"enabled"`
		EnableColor bool   `yaml:"enable_color"`
		LogLevel    string `yaml:"log_level"`
	} `yaml:"console"`
}

// PrometheusConfig represents the configuration for Prometheus.
type PrometheusConfig struct {
	Path    string `yaml:"path"`
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
}

// ListenConfig represents the configuration for the listening address.
type ListenConfig struct {
	Address   string `yaml:"address"`
	Community string `yaml:"community"`
}

type Redis struct {
	Enabled  bool   `yaml:"enabled"`
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
	Channel  string `yaml:"channel"`
}

type ScriptHandlerConfig struct {
	Enabled       bool   `yaml:"enabled"`
	CountHandlers int    `yaml:"count_handlers"`
	QueueSize     int    `yaml:"queue_size"`
	Command       string `yaml:"command"`
}
type Configuration struct {
	Logger        LoggerConfig        `yaml:"logger"`
	Prometheus    PrometheusConfig    `yaml:"prometheus"`
	Listen        ListenConfig        `yaml:"listen"`
	Redis         Redis               `yaml:"redis"`
	ScriptHandler ScriptHandlerConfig `yaml:"script_handler"`
}

func LoadConfig(path string, Config *Configuration) error {
	var err error
	var bytes []byte
	if path == "" {
		bytes, err = configFile.ReadFile("trap-listener.yml")
	} else {
		bytes, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return err
	}
	yamlConfig := string(bytes)
	expandedContent := ExpandEnvDefault(yamlConfig)

	err = yaml.Unmarshal([]byte(expandedContent), Config)
	fmt.Printf(`Loaded configuration from %v with env readed:
%v
`, path, expandedContent)
	if err != nil {
		return err
	}
	ConfigureLogger(Config)
	return nil
}

func ExpandEnvDefault(s string) string {
	return os.Expand(s, func(key string) string {
		// Определение дефолтного значения через двоеточие.
		parts := strings.SplitN(key, ":", 2)
		if len(parts) == 2 {
			if value, ok := os.LookupEnv(parts[0]); ok {
				return value
			}
			return parts[1]
		}
		return os.Getenv(key)
	})
}

func ConfigureLogger(conf *Configuration) {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05",
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	level, err := log.ParseLevel(conf.Logger.Console.LogLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(level)
}
