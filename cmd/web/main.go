package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const WEB_PORT = "8080"

func main() {
	db := initDB()
	fmt.Println(db)

	// create session
	session := initSession()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	wg := sync.WaitGroup{}

	app := Config{
		Session:  session,
		DB:       db,
		Wait:     &wg,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}
	// add safe exit.
	go safeExit(&app)
	app.serve()
}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", WEB_PORT),
		Handler: app.routes(),
	}
	app.InfoLog.Println("starting web server...")
	err := srv.ListenAndServe()
	fmt.Println(err, "err++++")
	if err != nil {
		log.Panic(err)
	}
}

func initDB() *sql.DB {
	dsn := os.Getenv("DSN")
	fmt.Println(dsn)
	connection, err := openDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	err = connection.Ping()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("connect success")
	}
	return connection
}

func openDB(dsn string) (*sql.DB, error) {
	const COUNT = 3
	var connection *sql.DB
	var err error
	for i := 0; i < COUNT; i++ {
		connection, err = sql.Open("pgx", dsn)
		if err != nil {
			if i < COUNT-1 {
				time.Sleep(time.Second * 1)
			}
			continue
		}
		break
	}
	return connection, err
}

func initSession() *scs.SessionManager {
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true
	return session
}

func initRedis() *redis.Pool {
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
	return redisPool
}

func safeExit(app *Config) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	app.Wait.Wait()
	fmt.Println("before exit")
	os.Exit(0)
}
