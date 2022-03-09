package main

import (
	"github.com/djumanoff/amqp"
	lib "github.com/kirigaikabuto/recommendation-main-store"
	setdata_common "github.com/kirigaikabuto/setdata-common"
	"log"
)

var (
	postgresUser         = "setdatauser"
	postgresPassword     = "123456789"
	postgresDatabaseName = "recommendation_system"
	postgresHost         = "localhost"
	postgresPort         = 5432
	postgresParams       = "sslmode=disable"
	amqpUrl              = "amqp://localhost:5672"
)

func main() {
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
		panic(err)
		return
	}
	scoreService := lib.NewScoreService(scoreStore)
	scoreCommandHandler := setdata_common.NewCommandHandler(scoreService)
	scoreAmqpEndpoints := lib.NewScoreAmqpEndpoints(scoreCommandHandler)
	//users
	usersStore, err := lib.NewPostgresUsersStore(config)
	if err != nil {
		panic(err)
		return
	}
	usersService := lib.NewUserService(usersStore)
	usersCommandHandler := setdata_common.NewCommandHandler(usersService)
	usersAmqpEndpoints := lib.NewUserAmqpEndpoints(usersCommandHandler)
	//movies
	movieStore, err := lib.NewMoviesPostgreStore(config)
	if err != nil {
		log.Fatal(err)
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
		panic(err)
		return
	}
	srv, err := sess.Server(serverConfig)
	if err != nil {
		panic(err)
		return
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
		panic(err)
		return
	}
}
