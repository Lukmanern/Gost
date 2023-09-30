package env

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	PublicKey  *[]byte
	PrivateKey *[]byte

	envFile *string
	cfg     Config

	PublicKeyReadOne  sync.Once
	PrivateKeyReadOne sync.Once
	cfgOnce           sync.Once
)

type Config struct {
	Title             string        `env-default:"Golang Test Fatur Rahman"`
	AccessTokenTTL    time.Duration `env:"ACCESS_TOKEN_TTL" env-default:"1440m" env-upd`
	Port              int           `env:"PORT" env-default:"7007"`
	SecretBytes       string        `env:"SECRET_BYTES" env-default:"secret"`
	MysqlRootHost     string        `env:"MYSQL_ROOT_HOST" env-upd`
	MysqlRootPassword string        `env:"MYSQL_ROOT_PASSWORD" env-upd`
	MysqlPort         string        `env:"MYSQL_PORT" env-upd`
	MysqlUser         string        `env:"MYSQL_USER" env-upd`
	MysqlPassword     string        `env:"MYSQL_PASSWORD" env-upd`
	MysqlDbname       string        `env:"MYSQL_DBNAME" env-upd`
	DatabaseURI       string
	PublicKey         string `env:"PUBLIC_KEY" env-required`
	PrivateKey        string `env:"PRIVATE_KEY" env-required`
	SMTPServer        string `env:"SMTP_SERVER" env-required`
	SMTPPort          int    `env:"SMTP_PORT" env-required`
	SMTPEmail         string `env:"SMTP_EMAIL" env-required`
	SMTPPassword      string `env:"SMTP_PASSWORD" env-required`
	ClientURL         string `env:"CLIENT_URL" env-required`
}

func ReadConfig(file string) *Config {
	cfgOnce.Do(func() {
		envFile = &file
		log.Printf(`Reading config file: "%s"`, *envFile)
		err := cleanenv.ReadConfig(file, &cfg)
		if err != nil {
			err := cleanenv.ReadEnv(&cfg)
			if err != nil {
				log.Fatalf("Config error %s", err.Error())
			}
		}
	})
	return &cfg
}
