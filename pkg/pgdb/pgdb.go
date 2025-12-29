package pgdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jdetok/bball-etl-cli/pkg/conn"
	_ "github.com/lib/pq"
)

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	ConnStr  string
}

type DBConfig struct {
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
}

func NewDBConf(maxOpen, maxIdle int, maxLife time.Duration) *DBConfig {
	return &DBConfig{
		MaxOpenConns: maxOpen,
		MaxIdleConns: maxIdle,
		ConnMaxLife:  maxLife,
	}
}

func NewPG(e *conn.DBEnv) *Postgres {
	return &Postgres{
		Host:     e.Host,
		Port:     e.Port,
		User:     e.User,
		Password: e.Pass,
		Database: e.Database,
	}
}

func NewPGConn(e *conn.DBEnv, conf *DBConfig) (*sql.DB, error) {
	pg := NewPG(e)
	pg.MakeConnStr()
	// fmt.Println(pg.ConnStr)
	db, err := pg.Conn()
	if err != nil {
		return nil, fmt.Errorf("error connecting to postgres\n%v", err)
	}
	if conf == nil {
		return db, nil
	}

	// set max connections
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	db.SetConnMaxLifetime(conf.ConnMaxLife)
	return db, nil
}

func (pg *Postgres) MakeConnStr() {
	pg.ConnStr = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pg.Host, pg.Port, pg.User, pg.Password, pg.Database)
}

func (pg *Postgres) Conn() (*sql.DB, error) {
	pg.MakeConnStr()
	db, err := sql.Open("postgres", pg.ConnStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf(
			"error pining postgres after successful conn: %e", err)
	}
	return db, err
}
