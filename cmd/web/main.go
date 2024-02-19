package main

import (
	"context"
	"flag"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"os"
	"time"
)

type application struct {
	cfg           config
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
	//models        data.Models
	infoLog  *log.Logger
	errorLog *log.Logger
}

type config struct {
	port int
	db   struct {
		dsn         string
		maxOpenConn int
		maxIdleTime string
	}
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "server port")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("FORUM_DB_DSN"), "PostgreSQL DSN")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle time")
	flag.IntVar(&cfg.db.maxOpenConn, "db-max-open-conn", 25, "PostgreSQL max open connections")

	flag.Parse()
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(cfg)
	if err != nil {
		errorLog.Println(err)
		return
	}
	defer db.Close()
	infoLog.Println("database connection established")
	templateCache, err := NewTemplateCache()
	if err != nil {
		errorLog.Println(err)
		return
	}
	formDecoder := form.NewDecoder()
	app := &application{
		cfg:           cfg,
		infoLog:       infoLog,
		errorLog:      errorLog,
		formDecoder:   formDecoder,
		templateCache: templateCache,
		//models:        data.NewModels(db),
	}
	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

func openDB(cfg config) (*pgxpool.Pool, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	dbConfig.MaxConnIdleTime = duration
	dbConfig.MaxConns = int32(cfg.db.maxOpenConn)
	db, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
