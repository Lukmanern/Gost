package env

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppName           string        `env:"APP_NAME"`
	AppInProduction   bool          `env:"APP_IN_PRODUCTION"`
	AppKey            string        `env:"APP_SECRET_KEY"`
	AppAccessTokenTTL time.Duration `env:"APP_ACCESS_TOKEN_TTL"`
	AppPort           int           `env:"APP_PORT"`

	DatabaseRootHost     string `env:"DB_HOST"`
	DatabaseRootPassword string `env:"DB_ROOT_PASSWORD"`
	DatabasePort         string `env:"DB_PORT"`
	DatabaseUser         string `env:"DB_USERNAME"`
	DatabasePassword     string `env:"DB_PASSWORD"`
	DatabaseName         string `env:"DB_DATABASE"`
	DatabaseURI          string

	PublicKey  string `env:"PUBLIC_KEY"`
	PrivateKey string `env:"PRIVATE_KEY"`

	SMTPServer   string `env:"SMTP_SERVER"`
	SMTPPort     int    `env:"SMTP_PORT"`
	SMTPEmail    string `env:"SMTP_EMAIL"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	ClientURL    string `env:"CLIENT_URL"`
}

var (
	PublicKey  *[]byte
	PrivateKey *[]byte

	envFile *string
	cfg     Config

	PublicKeyReadOne  sync.Once
	PrivateKeyReadOne sync.Once
	cfgOnce           sync.Once
)

func ReadConfig(filePath string) *Config {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Panicf(`.env file isn't exist/found: "%s"`, filePath)
	}
	cfgOnce.Do(func() {
		envFile = &filePath
		log.Printf(`Reading config file: "%s"`, *envFile)
		err := cleanenv.ReadConfig(filePath, &cfg)
		if err != nil {
			err := cleanenv.ReadEnv(&cfg)
			if err != nil {
				log.Fatalf("Config error %s", err.Error())
			}
		}
	})
	return &cfg
}

func Configuration() Config {
	if envFile == nil {
		log.Panic(`configuration file is not set. Call ReadConfig("path_to_file") first`)
	}
	err := cleanenv.UpdateEnv(&cfg)
	if err != nil {
		log.Fatalf("Config error %s", err.Error())
	}
	return cfg
}

func (c Config) GetDatabaseURI() string {
	c.DatabaseURI = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&multiStatements=true&parseTime=true",
		c.DatabaseUser, c.DatabasePassword, c.DatabaseRootHost, c.DatabasePort, c.DatabaseName)

	return c.DatabaseURI
}

func (c Config) ShowConfig() {
	fmt.Printf("%-21s: %s\n", "AppName", c.AppName)
	fmt.Printf("%-21s: %v\n", "AppInProduction", c.AppInProduction)
	fmt.Printf("%-21s: %s\n", "AppKey", c.AppKey)
	fmt.Printf("%-21s: %s\n", "AppAccessTokenTTL", c.AppAccessTokenTTL)
	fmt.Printf("%-21s: %d\n", "AppPort", c.AppPort)

	fmt.Printf("%-21s: %s\n", "DatabaseRootHost", c.DatabaseRootHost)
	fmt.Printf("%-21s: %s\n", "DatabaseRootPassword", c.DatabaseRootPassword)
	fmt.Printf("%-21s: %s\n", "DatabasePort", c.DatabasePort)
	fmt.Printf("%-21s: %s\n", "DatabaseUser", c.DatabaseUser)
	fmt.Printf("%-21s: %s\n", "DatabasePassword", c.DatabasePassword)
	fmt.Printf("%-21s: %s\n", "DatabaseName", c.DatabaseName)
	fmt.Printf("%-21s: %s\n", "DatabaseURI", c.DatabaseURI)

	// add more as needed
}
