package main

import (
	"errors"
	"flag"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fraenkel/candiedyaml"
	"github.com/garyburd/redigo/redis"

	proleroutes "github.com/winston-ci/prole/routes"

	"github.com/winston-ci/winston/api"
	apiroutes "github.com/winston-ci/winston/api/routes"
	"github.com/winston-ci/winston/builder"
	"github.com/winston-ci/winston/config"
	"github.com/winston-ci/winston/db"
	"github.com/winston-ci/winston/endpoint"
	"github.com/winston-ci/winston/server"
)

var configPath = flag.String(
	"config",
	"",
	"path to winston server config .yml",
)

var proleURL = flag.String(
	"proleURL",
	"http://127.0.0.1:4637",
	"address denoting the prole service",
)

var redisAddr = flag.String(
	"redisAddr",
	"127.0.0.1:6379",
	"redis server address",
)

var peerURL = flag.String(
	"peerURL",
	"http://127.0.0.1:8081",
	"external URL of the api server",
)

var listenAddr = flag.String(
	"listenAddr",
	":8080",
	"port for the web server to listen on",
)

var apiListenAddr = flag.String(
	"apiListenAddr",
	":8081",
	"port for the api to listen on",
)

func main() {
	flag.Parse()

	if *configPath == "" {
		fatal(errors.New("must specify -config"))
	}

	configFile, err := os.Open(*configPath)
	if err != nil {
		fatal(err)
	}

	var config config.Config
	err = candiedyaml.NewDecoder(configFile).Decode(&config)
	if err != nil {
		fatal(err)
	}

	configFile.Close()

	redisDB := db.NewRedis(redis.NewPool(func() (redis.Conn, error) {
		return redis.DialTimeout("tcp", *redisAddr, 5*time.Second, 0, 0)
	}, 20))

	winstonApiURL, err := url.Parse(*peerURL)
	if err != nil {
		fatal(err)
	}

	winstonEndpoint := endpoint.EndpointRoutes{
		URL:    winstonApiURL,
		Routes: apiroutes.Routes,
	}

	proleEarl, err := url.Parse(*proleURL)
	if err != nil {
		fatal(err)
	}

	proleEndpoint := endpoint.EndpointRoutes{
		URL:    proleEarl,
		Routes: proleroutes.Routes,
	}

	builder := builder.NewBuilder(redisDB, proleEndpoint, winstonEndpoint)

	serverHandler, err := server.New(config, redisDB, builder)
	if err != nil {
		fatal(err)
	}

	apiHandler, err := api.New(redisDB)
	if err != nil {
		fatal(err)
	}

	errs := make(chan error, 2)

	go func() {
		errs <- http.ListenAndServe(*listenAddr, serverHandler)
	}()

	go func() {
		errs <- http.ListenAndServe(*apiListenAddr, apiHandler)
	}()

	fatal(<-errs)
}

func fatal(err error) {
	println(err.Error())
	os.Exit(1)
}