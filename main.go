package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type postgresConfig struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     string `envconfig:"PORT" required:"true"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	DBName   string `envconfig:"DB_NAME" required:"true"`
	SSLMode  string `envconfig:"SSL_MODE" default:"disable"`
	TimeZone string `envconfig:"TIME_ZONE" required:"true"`
}

func (c postgresConfig) toDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", c.Host, c.User, c.Password, c.DBName, c.Port, c.SSLMode, c.TimeZone)
}

func newPostgresConfig() (postgresConfig, error) {
	env, ok := os.LookupEnv("ENV")
	if ok && env != "" {
		if err := godotenv.Load(); err != nil {
			return postgresConfig{}, err
		}
	}

	var cfg postgresConfig
	envconfig.MustProcess("POSTGRES", &cfg)
	return cfg, nil
}

type migrateLogger struct {
	logger *zap.Logger
}

func (l migrateLogger) Printf(format string, v ...interface{}) {
	l.logger.Sugar().Infof(format, v)
}

func (l migrateLogger) Verbose() bool {
	return true
}

func main() {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	config, err := newPostgresConfig()
	if err != nil {
		logger.Panic("error loading config", zap.Error(err))
	}

	db, err := sql.Open("postgres", config.toDSN())
	if err != nil {
		logger.Panic("error openning sql database", zap.Error(err))
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Panic("error creating driver", zap.Error(err))
	}

	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
	if err != nil {
		logger.Panic("error creating migration", zap.Error(err))
	}

	m.Log = migrateLogger{logger: logger}

	_ = m.Up()
}
