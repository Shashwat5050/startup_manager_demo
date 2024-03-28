package main

import (
	"flag"
	"log"
	"startup-manager/config"
	"startup-manager/controller"
	core "startup-manager/core/config"
	coreLogger "startup-manager/core/logger"
	postgres "startup-manager/core/postgres"
	"startup-manager/usecase"
	database "startup-manager/usecase/repository"
	"sync"

	nomadapi "startup-manager/core/nomad"

	"go.uber.org/zap"
)

func main() {
	var configFile string

	flag.StringVar(&configFile, "c", "config.yml", "config file")
	flag.Parse()
	logger, err := coreLogger.NewDefaultLogger()
	if err != nil {
		log.Fatalf("cannot initialize otel logger: %v", err)
	}
	logger.Debug("initialized logger")
	conf, err := core.LoadConfig[*config.Config](configFile)
	if err != nil {
		logger.Error("cannot load config",
			zap.Error(err),
			zap.String("filename", configFile))
	}
	logger.Debug("config loaded", zap.String("filename", configFile))
	dbConfig := conf.GetDbConfig()
	logger.Debug("dbconfig is", zap.Any("dbconfig", dbConfig))

	pg, err := postgres.NewPostgres(dbConfig)

	if err != nil {
		logger.Error("cannot initialize postgres", zap.Error(err))
		panic(err)
	}
	defer pg.Close()
	logger.Info("postgres initialized")

	startupRepo := database.NewStartupRepository(pg)
	logger.Info("repository initialized")

	// initialize nomad client
	nomadClient, err := nomadapi.NewNomadClient(conf.NomadURL)

	if err != nil {
		logger.Error("cannot initialize nomad client", zap.Error(err), zap.String("url", conf.NomadURL))
		panic(err)
	}
	startupUsecase := usecase.NewStartUpUsecase(logger, startupRepo, nomadClient)

	logger.Info("usecase initialized", zap.Any("usecase", startupUsecase))

	startupController := controller.NewStartupController(logger, startupUsecase)
	logger.Info("controller initialized")

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		logger.Info("starting http server")

		if err := startupController.Start(); err != nil {
			logger.Error("cannot start http server", zap.Error(err))
			panic(err)
		}

	}()

	wg.Wait()
	logger.Info("startup manager server closed")
}
