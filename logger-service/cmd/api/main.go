package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/codercollo/micro/logger-service/data"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gPRCPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// Connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// context to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// close the connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// Register RPC server
	err = rpc.Register(new(RPCServer))

	go app.rpcListen()

	go app.gRPCListen()

	// start web server
	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: app.routes(),
	}

	log.Printf("Logger service starting on port %s", webPort)

	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server failed: %v", err)
	}
}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port", rpcPort)

	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeConn(rpcConn)

	}
}

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// connect
	c, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Println("Error connecting", err)
		return nil, err
	}

	return c, nil
}
