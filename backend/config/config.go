package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
)

// Global env var spreads across every package
// Not ideal, should be modified later
// https://www.alexedwards.net/blog/organising-database-access
var DB *sql.DB
var Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}).With().Timestamp().Logger()
var Conf ConfStruct

type ConfServer struct {
	Port         int           `env:"SERVER_PORT,required" envDefault:"8080"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
	LogLevel     string        `env:"SERVER_LOG_LEVEL,required"`
}

type ConfDB struct {
	Host     string `env:"DB_HOST,required"`
	Port     int    `env:"DB_PORT,required"`
	DBName   string `env:"DB_NAME,required"`
	Username string `env:"DB_USER,required"`
	Password string `env:"DB_PASS,required"`
}

type ConfPowens struct {
	ClientId       string   `env:"POWENS_CLIENT_ID,required"`
	ClientSecret   string   `env:"POWENS_CLIENT_SECRET,required"`
	Domain         string   `env:"POWENS_DOMAIN,required"`
	WebviewUrl     string   `env:"POWENS_WEBVIEW_URL,required"`
	RedirectUrl    string   `env:"POWENS_REDIRECT_URL,required"`
	WhitelistedIPs []string `env:"POWENS_WHITELISTED_IPS,required"`
}

type ConfOther struct {
	Language string `env:"OTHER_LANGUAGE,required"`
}

type ConfStruct struct {
	Server ConfServer
	DB     ConfDB
	Powens ConfPowens
	Other  ConfOther
}

func Init() {

	var err error

	// Read env values from .env. Remove these 3 lines if your envs are exported from somewhere else, like Dockerfile for example
	// These lines are usefull for local dev only and should not be used in production

	// err = godotenv.Load()
	// if err != nil {
	// 	Logger.Fatal().Err(err).Msg("Failed to load .env file")
	// }

	if err := env.Parse(&Conf.DB); err != nil {
		Logger.Fatal().Err(err).Msg("Failed to load env for DB")
	}
	if err := env.Parse(&Conf.Server); err != nil {
		Logger.Fatal().Err(err).Msg("Failed to load env for server")
	}
	if err := env.Parse(&Conf.Powens); err != nil {
		Logger.Fatal().Err(err).Msg("Failed to load env for Powens")
	}
	if err := env.Parse(&Conf.Other); err != nil {
		Logger.Fatal().Err(err).Msg("Failed to load env for Other")
	}

	// Set log level according to env value SERVER_LOG_LEVEL
	switch Conf.Server.LogLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		Logger.Fatal().Msgf("Unsupported value '%s' for SERVER_LOG_LEVEL. Should be trace, debug, info, warn, error, fatal or panic", Conf.Server.LogLevel)
	}

	// Capture connection properties and connect to DB.
	cfg := mysql.Config{
		User:   Conf.DB.Username,
		Passwd: Conf.DB.Password,
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%d", Conf.DB.Host, Conf.DB.Port),
		DBName: Conf.DB.DBName,
	}

	DB, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		Logger.Fatal().Err(err).Msg("")
	}

	err = DB.Ping()
	if err != nil {
		Logger.Fatal().Err(err).Msg("")
	}

	Logger.Info().Msg("Successfully ping DB")
}
