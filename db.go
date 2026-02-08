package db

import (
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/utils"
)

const (
	MysqlDriver    = "mysql"
	SqliteDriver   = "sqlite"
	PostgresDriver = "postgres"
)

func NewDB(dialect gorm.Dialector, opts ...gorm.Option) (Provider, error) {
	db, err := gorm.Open(dialect, opts...)
	if err != nil {
		return nil, err
	}
	return &provider{db: db}, nil
}

func NewDBFrom(cfg *Config, opts ...gorm.Option) (Provider, error) {
	if cfg == nil {
		return nil, errors.New("db config can't be nil")
	}

	if cfg.Driver == MysqlDriver && !utils.Contains(mysql.UpdateClauses, "RETURNING") {
		mysql.UpdateClauses = append(mysql.UpdateClauses, "RETURNING")
	}
	opts = append(opts, &gorm.Config{TranslateError: true})

	db, err := newDBFrom(cfg, opts...)
	if err != nil {
		return nil, err
	}

	return &provider{db: db}, nil
}

func newDBFrom(cfg *Config, opts ...gorm.Option) (*gorm.DB, error) {
	var err error
	var d *gorm.DB

	switch cfg.Driver {
	case MysqlDriver:
		d, err = gorm.Open(mysql.Open(cfg.DSN), opts...)
	case SqliteDriver:
		d, err = gorm.Open(sqlite.Open(cfg.DSN), opts...)
	case PostgresDriver:
		d, err = gorm.Open(postgres.Open(cfg.DSN), opts...)
	default:
		err = fmt.Errorf("database %q is not support", cfg.Driver)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open database, error: %v", err)
	}
	if d == nil {
		return nil, fmt.Errorf("failed to open database")
	}

	db, err := d.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect database, error: %v", err)
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return d, nil
}
