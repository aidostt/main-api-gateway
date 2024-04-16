package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

const (
	defaultHTTPPort               = "8000"
	defaultHTTPRWTimeout          = 10 * time.Second
	defaultHTTPMaxHeaderMegabytes = 1
	defaultAccessTokenTTL         = 15 * time.Minute
	defaultRefreshTokenTTL        = 12 * time.Hour
	defaultGRPCPort               = "443"
	authority                     = "api-gateway"
	EnvLocal                      = "local"
)

type (
	Config struct {
		Environment  string
		Authority    string
		GRPC         GRPCConfig         `mapstructure:"grpc"`
		Users        MicroserviceConfig `mapstructure:"userMicroservice"`
		Reservations MicroserviceConfig `mapstructure:"reservationMicroservice"`
		HTTP         HTTPConfig         `mapstructure:"http"`
		JWT          JWTConfig          `mapstructure:"jwt"`
	}

	GRPCConfig struct {
		Host    string        `mapstructure:"host"`
		Port    string        `mapstructure:"port"`
		Timeout time.Duration `mapstructure:"timeout"`
	}
	JWTConfig struct {
		AccessTokenTTL  time.Duration `mapstructure:"accessTokenTTL"`
		RefreshTokenTTL time.Duration `mapstructure:"refreshTokenTTL"`
		SigningKey      string
	}
	MicroserviceConfig struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	}
	HTTPConfig struct {
		Host               string        `mapstructure:"host"`
		Port               string        `mapstructure:"port"`
		ReadTimeout        time.Duration `mapstructure:"readTimeout"`
		WriteTimeout       time.Duration `mapstructure:"writeTimeout"`
		MaxHeaderMegabytes int           `mapstructure:"maxHeaderBytes"`
	}
)

func Init(configsDir, envDir string) (*Config, error) {
	populateDefaults()
	loadEnvVariables(envDir)
	if err := parseConfigFile(configsDir); err != nil {
		return nil, err
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	setFromEnv(&cfg)

	return &cfg, nil
}

func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("http", &cfg.HTTP); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("userMicroservice", &cfg.Users); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("reservationMicroservice", &cfg.Reservations); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("jwt", &cfg.JWT); err != nil {
		return err
	}
	return viper.UnmarshalKey("grpc", &cfg.GRPC)
}

func setFromEnv(cfg *Config) {
	cfg.HTTP.Host = os.Getenv("HTTP_HOST")
	cfg.GRPC.Host = os.Getenv("GRPC_HOST")
	cfg.Environment = EnvLocal
	cfg.Authority = authority
	cfg.JWT.SigningKey = os.Getenv("JWT_SIGNING_KEY")
}

func parseConfigFile(folder string) error {
	viper.AddConfigPath(folder)
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.MergeInConfig()
}

func loadEnvVariables(envPath string) {
	err := godotenv.Load(envPath)

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

}

func populateDefaults() {
	viper.SetDefault("http.port", defaultHTTPPort)
	viper.SetDefault("grpc.port", defaultGRPCPort)
	viper.SetDefault("http.max_header_megabytes", defaultHTTPMaxHeaderMegabytes)
	viper.SetDefault("http.timeouts.read", defaultHTTPRWTimeout)
	viper.SetDefault("http.timeouts.write", defaultHTTPRWTimeout)
	viper.SetDefault("jwt.accessTokenTTL", defaultAccessTokenTTL)
	viper.SetDefault("jwt.refreshTokenTTL", defaultRefreshTokenTTL)
}
