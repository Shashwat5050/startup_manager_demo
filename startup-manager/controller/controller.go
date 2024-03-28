package controller

import (
	"net/http"
	"startup-manager/core/logger"
	"startup-manager/usecase"

	"github.com/gin-gonic/gin"
)

type StartupController struct {
	logger  logger.Logger
	usecase *usecase.StartUpUsecase
	httpMux *http.ServeMux
}

func NewStartupController(logger logger.Logger, usecase *usecase.StartUpUsecase) *StartupController {
	return &StartupController{
		logger:  logger,
		usecase: usecase,
		httpMux: http.NewServeMux(),
	}
}

func (sc *StartupController) registerRoutes() {
	router := gin.Default()
	startupRoute := router.Group("/")

	startupRoute.POST("/addstartup", sc.AddStartupHandler)
	startupRoute.GET("/getstartup",sc.GetStartup)
	startupRoute.DELETE("/deletestartup",sc.DeleteStartupInfo)
	startupRoute.GET("/getDefaultParameters",sc.GetGameEnvironments)
	startupRoute.GET("/get_game_info",sc.GetGameInfo)
	startupRoute.GET("/get_default_command",sc.GetDefaultStartupCommand)
	sc.httpMux.Handle("/", router)

}

func (sc *StartupController) Start() error {
	sc.registerRoutes()
	server := http.Server{
		Handler: sc.httpMux,
		Addr:    ":6000",
	}
	return server.ListenAndServe()
}
