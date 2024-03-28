package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"startup-manager/core/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (sc *StartupController) AddStartupHandler(ctx *gin.Context) {
	var startupRequest StartupInfoRequest

	if err := ctx.BindJSON(&startupRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println(startupRequest)
	startupInfo := models.StartupInfo{
		ServerID:      startupRequest.ServerID,
		Variables:     startupRequest.Variables,
		CreatedAt:     startupRequest.CreatedAt,
		UpdatedAt:     &startupRequest.UpdatedAt,
		DeletedAt:     &startupRequest.DeletedAt,
	}
	startup_id, err := sc.usecase.AddStartup(ctx, &startupInfo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"startup_id": startup_id, "message": "startup added successfully"})

}

func(sc *StartupController)GetStartup(ctx *gin.Context){
	id:=ctx.Query("startup_id")

	startup,err:=sc.usecase.GetStartup(ctx,id)
	if err!=nil{
		ctx.JSON(http.StatusBadRequest,gin.H{"error":err})
		return
	}
	ctx.JSON(http.StatusOK,gin.H{"startup":startup})
	return
}

func(sc *StartupController)DeleteStartupInfo(ctx *gin.Context){
	id:=ctx.Query("startup_id")

	err:=sc.usecase.DeleteStartupInfo(ctx,id)
	if err!=nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":"startup not found"})
		return
	}
	ctx.JSON(http.StatusOK,gin.H{"status":"startup info deleted"})
	
}
func(sc *StartupController)GetGameEnvironments(ctx *gin.Context){
	name:=ctx.Query("game_name")

	envs,err:=sc.usecase.GetGameEnvironments(ctx,name)
	if err!=nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":"env not found"})
		return
	}
	ctx.JSON(http.StatusOK,gin.H{"envs":envs})
}

func(sc *StartupController)GetGameInfo(ctx *gin.Context){
	game:=ctx.Query("game")

	game_info,err:=sc.usecase.GetGameInfo(ctx,game)
	if err!=nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":"game not found"})
		return
	}
	ctx.JSON(http.StatusOK,gin.H{"game":game_info})

}
func(sc *StartupController)GetDefaultStartupCommand(ctx *gin.Context){
	game:=ctx.Query("game")

	command,err:=sc.usecase.GetDefaultStartupCommand(ctx,game)
	if err!=nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":"command not found"})
		return
	}
	ctx.JSON(http.StatusOK,gin.H{"command":command})
}

func convertMapToJSON(data map[string]interface{}) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Handle error, possibly log it
		return []byte("{}") // Default to an empty JSON object in case of error
	}
	return jsonData
}

type StartupInfoRequest struct {
	ServerID      uuid.UUID              `json:"server_id"`
	StartupParams map[string]interface{} `json:"startup_parameters"`
	Variables     map[string]interface{} `json:"variables"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	DeletedAt     time.Time              `json:"deleted_at"`
}
