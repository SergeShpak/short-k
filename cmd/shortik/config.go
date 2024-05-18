package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"dario.cat/mergo"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"

	"shortik/internal/core/app"
	"shortik/internal/infra/api/rest"
	"shortik/internal/infra/store/db"
)

//nolint:govet // fieldalignement check is irrelevant heree
type Config struct {
	App     app.ConfigParams         `yaml:"app"`
	DB      db.ConfigParams          `yaml:"-"`
	HTTP    rest.ServerConfigParams  `yaml:"http"`
	Handler rest.HandlerConfigParams `yaml:"handler"`
	Run     RunConfig                `yaml:"run"`
}

type RunConfig struct {
	HTTPServerShutdownTimeout time.Duration `yaml:"httpServerShutdownTimeout" validate:"required,gt=0"`
	DBCloseTimeoout           time.Duration `yaml:"dbCloseTimeout" validate:"required,gt=0"`
	ShutdownTimeout           time.Duration `yaml:"shutdownTimeout" validate:"required,gt=0"`
}

func getDefaultRunConfig() RunConfig {
	return RunConfig{
		HTTPServerShutdownTimeout: time.Second * 30,
		DBCloseTimeoout:           time.Second * 30,
		ShutdownTimeout:           time.Second * 60,
	}
}

func GetConfig() (Config, error) {
	var cfg Config

	flags, err := getFlags(os.Args[1:])
	if err != nil {
		return cfg, fmt.Errorf("failed to get flags: %w", err)
	}

	cfgFile, err := os.ReadFile(flags.ConfigFile)
	if err != nil {
		return cfg, fmt.Errorf("failed to read the file \"%s\": %w", flags.ConfigFile, err)
	}

	cfg, err = getYAMLConfig(cfgFile, flags)
	if err != nil {
		return cfg, fmt.Errorf("failed to get the YAML configuration: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return cfg, fmt.Errorf("config validation has failed: %w", err)
	}

	return cfg, nil
}

type flags struct {
	DSN        string
	ConfigFile string
}

func getFlags(args []string) (flags, error) {
	fs := flag.NewFlagSet("fs", flag.ContinueOnError)
	var outputBuf bytes.Buffer
	fs.SetOutput(&outputBuf)

	f := flags{}
	fs.StringVar(&f.DSN, "dsn", "", "DB connection string")
	fs.StringVar(&f.ConfigFile, "config", "", "configuration file path")
	if err := fs.Parse(args); err != nil {
		return f, fmt.Errorf("failed to parse flags: %w", err)
	}

	if err := validateFlags(f); err != nil {
		fs.PrintDefaults()
		return f, fmt.Errorf("%w\n\nUsage:\n%s", err, outputBuf.String())
	}

	return f, nil
}

func validateFlags(f flags) error {
	if len(f.DSN) == 0 {
		return errors.New("DB connection string must be specified")
	}
	if len(f.ConfigFile) == 0 {
		return errors.New("path to the configuration file must be specified")
	}
	return nil
}

func getDefaultConfig() Config {
	return Config{
		App:     app.GetDefaultConfigParams(),
		DB:      db.GetDefaultConfigParams(),
		HTTP:    rest.GetDefaultServerConfigParams(),
		Handler: rest.GetDefaultHandlerConfigParams(),
		Run:     getDefaultRunConfig(),
	}
}

func getYAMLConfig(data []byte, flags flags) (Config, error) {
	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return fileCfg, fmt.Errorf("failed to unmarshal the YAML config: %w", err)
	}
	fileCfg.DB.DSN = flags.DSN

	cfg := getDefaultConfig()

	if err := mergo.Merge(&cfg, fileCfg, mergo.WithOverride); err != nil {
		return cfg, fmt.Errorf("failed to merge default and custom configurations: %w", err)
	}

	return cfg, nil
}
