package config

import "github.com/joeshaw/envdecode"

type DatabaseConfig struct {
	Name              string `env:"DB_NAME"`
	Port              string `env:"DB_PORT"`
	Host              string `env:"DB_HOST"`
	Username          string `env:"DB_USERNAME"`
	Password          string `env:"DB_PASSWORD"`
	MaxOpenConnection int    `env:"max_open_connection"`
	MaxIdleConnection int    `env:"max_idle_connection"`
}

type S3Config struct {
	ID        string `env:"S3_ID"`
	SecretKey string `env:"S3_SECRET_KEY"`
	Bucket    string `env:"S3_BUCKET_NAME"`
	Region    string `env:"S3_REGION"`
}

type Config struct {
	Database          DatabaseConfig
	AppPort           string `env:"APP_PORT"`
	PrometheusAddress string `env:"PROMETHEUS_ADDRESS"`
	Env               string `env:"ENV"`

	// security-related options
	JWTSecret  string `env:"JWT_SECRET"`
	BcryptSalt int    `env:"BCRYPT_SALT"`

	// S3Enabled is a flag which if set to true, will set image upload to s3
	S3Enabled bool `env:"S3_ENABLED"`

	// S3 stores config to connect to S3
	S3 S3Config
}

func InitializeConfig() Config {
	var cfg Config
	err := envdecode.Decode(&cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
