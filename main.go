package main

import (
	"fmt"
	"github.com/djumanoff/amqp"
	"github.com/joho/godotenv"
	lib "github.com/kirigaikabuto/recommendation-main-store"
	setdata_common "github.com/kirigaikabuto/setdata-common"
	"github.com/urfave/cli"
	"log"
	"os"
	"strconv"
)

var (
	configPath           = ""
	postgresUser         = "setdatauser"
	postgresPassword     = "123456789"
	postgresDatabaseName = "recommendation_system"
	postgresHost         = "localhost"
	postgresPort         = 5432
	postgresParams       = "sslmode=disable"
	amqpUrl              = ""
	flags                = []cli.Flag{
		&cli.StringFlag{
			Name:        "config, c",
			Usage:       "path to .env config file",
			Destination: &configPath,
		},
	}
)

func parseEnvFile() {
	// Parse config file (.env) if path to it specified and populate env vars
	fmt.Println(configPath)
	if configPath != "" {
		godotenv.Overload(configPath)
	} else {
		godotenv.Overload("dev.env")
	}
	amqpUrl = os.Getenv("AMQP_URL")
	fmt.Println(amqpUrl)
	postgresUser = os.Getenv("postgresUser")
	postgresPassword = os.Getenv("postgresPassword")
	postgresDatabaseName = os.Getenv("postgresDatabaseName")
	postgresHost = os.Getenv("postgresHost")
	postgresPortStr := os.Getenv("postgresPort")
	postgresParams = os.Getenv("postgresParams")
	postgresPort, _ = strconv.Atoi(postgresPortStr)
	fmt.Println(postgresHost)
}

func main() {
	app := cli.NewApp()
	app.Name = "recommendation system work-api"
	app.Description = ""
	app.Usage = "recommendation system work-api"
	app.UsageText = "recommendation system work-api"
	app.Flags = flags
	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	parseEnvFile()
	config := lib.PostgresConfig{
		Host:     postgresHost,
		Port:     postgresPort,
		User:     postgresUser,
		Password: postgresPassword,
		Database: postgresDatabaseName,
		Params:   postgresParams,
	}
	//score
	scoreStore, err := lib.NewScorePostgreStore(config)
	if err != nil {
		return err
	}
	scoreService := lib.NewScoreService(scoreStore)
	scoreCommandHandler := setdata_common.NewCommandHandler(scoreService)
	scoreAmqpEndpoints := lib.NewScoreAmqpEndpoints(scoreCommandHandler)
	//users
	usersStore, err := lib.NewPostgresUsersStore(config)
	if err != nil {
		return err
	}
	usersService := lib.NewUserService(usersStore)
	usersCommandHandler := setdata_common.NewCommandHandler(usersService)
	usersAmqpEndpoints := lib.NewUserAmqpEndpoints(usersCommandHandler)
	//movies
	movieStore, err := lib.NewMoviesPostgreStore(config)
	if err != nil {
		return err
	}
	movieService := lib.NewMovieService(movieStore)
	moviesAmqpEndpoints := lib.NewAMQPEndpointFactory(movieService)

	rabbitConfig := amqp.Config{
		AMQPUrl:  amqpUrl,
		LogLevel: 5,
	}
	serverConfig := amqp.ServerConfig{
		ResponseX: "response",
		RequestX:  "request",
	}
	sess := amqp.NewSession(rabbitConfig)
	err = sess.Connect()
	if err != nil {
		return err
	}
	srv, err := sess.Server(serverConfig)
	if err != nil {
		return err
	}
	srv.Endpoint("score.create", scoreAmqpEndpoints.CreateScoreAmqpEndpoint())
	srv.Endpoint("score.list", scoreAmqpEndpoints.ListScoreAmqpEndpoint())

	srv.Endpoint("users.create", usersAmqpEndpoints.MakeCreateUserAmqpEndpoint())
	srv.Endpoint("users.get", usersAmqpEndpoints.MakeGetUserAmqpEndpoint())
	srv.Endpoint("users.list", usersAmqpEndpoints.MakeListUserAmqpEndpoint())
	srv.Endpoint("users.update", usersAmqpEndpoints.MakeUpdateUserAmqpEndpoint())
	srv.Endpoint("users.delete", usersAmqpEndpoints.MakeDeleteUserAmqpEndpoint())
	srv.Endpoint("users.getByUsernameAndPassword", usersAmqpEndpoints.MakeGetUserByUsernameAndPasswordAmqpEndpoint())

	srv.Endpoint("movie.get", moviesAmqpEndpoints.GetMovieByIdAMQPEndpoint())
	srv.Endpoint("movie.create", moviesAmqpEndpoints.CreateMovieAMQPEndpoint())
	srv.Endpoint("movie.list", moviesAmqpEndpoints.ListMoviesAMQPEndpoint())
	srv.Endpoint("movie.update", moviesAmqpEndpoints.UpdateProductAMQPEndpoint())
	srv.Endpoint("movie.delete", moviesAmqpEndpoints.DeleteMovieAMQPEndpoint())
	srv.Endpoint("movie.getByName", moviesAmqpEndpoints.GetMovieByNameAMQPEndpoint())
	err = srv.Start()
	if err != nil {
		return err
	}
	return nil
}
