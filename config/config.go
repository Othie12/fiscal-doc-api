package config

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

var (
	ServerConfig Server
	MysqlConfig  Mysql
	OracleConfig Oracle
	CorsConfig   cors.Config
)

func LoadConfig() {
	godotenv.Load(".env")
	var ctx context.Context

	err := envconfig.Process(ctx, &ServerConfig)
	if err != nil {
		log.Fatalln("Error loading env: " + err.Error())
	}

	//////////////////// MYSQL CONFIG ////////////////////
	err = envconfig.Process(ctx, &MysqlConfig)
	if err != nil {
		log.Fatalln("Error loading MysqlEnv: " + err.Error())
	} else {
		log.Println("Mysql Env loaded succesfuly.")
	}

	//////////////////// ORACLE CONFIG ////////////////////
	err = envconfig.Process(ctx, &OracleConfig)
	if err != nil {
		log.Fatalln("Error loading OracleEnv: " + err.Error())
	} else {
		log.Println("Oracle Env loaded succesfuly.")
	}

	//////////////////// CORS CONFIG ////////////////////
	CorsConfig = cors.DefaultConfig()
	CorsConfig.AllowAllOrigins = true
	CorsConfig.AllowMethods = []string{"POST", "GET", "PUT", "OPTIONS", "PATCH"}
	CorsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"}
	CorsConfig.ExposeHeaders = []string{"Content-Length"}
	CorsConfig.AllowCredentials = true
	CorsConfig.MaxAge = 12 * time.Hour
}

type Server struct {
	Name   string `env:"APP_NAME"`
	Port   string `env:"APP_PORT"`
	Secret string `env:"APP_SECRET"`
	Host   string `env:"APP_HOST"`
	Mode   string `env:"APP_ENV"`
}

type Mysql struct {
	Username string `env:"MYSQL_USERNAME"`
	Password string `env:"MYSQL_PASSWORD"`
	Host     string `env:"MYSQL_HOST"`
	Port     string `env:"MYSQL_PORT"`
	DBName   string `env:"MYSQL_DBNAME"`
}

type Oracle struct {
	Username string `env:"ORACLE_USERNAME"`
	Password string `env:"ORACLE_PASSWORD"`
	Host     string `env:"ORACLE_HOST"`
	Port     string `env:"ORACLE_PORT"`
	DBName   string `env:"ORACLE_DBNAME"`
}
