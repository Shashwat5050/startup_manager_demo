package usecase

import (
	"context"
	"errors"
	"fmt"
	"text/template"
	"log"
	"regexp"
	"startup-manager/core/logger"
	"startup-manager/core/models"
	nomadapi "startup-manager/core/nomad"
	"startup-manager/usecase/repository"
	"strings"

	"github.com/google/uuid"
)

type ServerParams struct {
	StartupCommand string
	Variables      map[string]interface{}
}

type StartUpUsecase struct {
	logger      logger.Logger
	repository  *repository.StartupRepository
	nomadClient *nomadapi.NomadClient
}

func NewStartUpUsecase(logger logger.Logger, repository *repository.StartupRepository, nomadClient *nomadapi.NomadClient) *StartUpUsecase {
	return &StartUpUsecase{
		logger:      logger,
		repository:  repository,
		nomadClient: nomadClient,
	}
}

func (su *StartUpUsecase) AddStartup(ctx context.Context, startup *models.StartupInfo) (string, error) {

	command, err := su.GetGameStartupCommand(ctx, startup.ServerID)
	if err != nil {
		return "", err
	}
	log.Println(command, "*********")
	filledcommand, err := generateStartupCommand(command, startup.Variables)
	if err != nil {
		return "", err
	}
	log.Println(filledcommand)
	startup.StartupCommand = filledcommand
	log.Println(startup.StartupCommand)
	startup_id, err := su.repository.AddStartupParams(ctx, startup)
	if err != nil {
		return "", err
	}
	log.Println(startup.StartupCommand)
	err = su.repository.UpdateGSCommand(ctx, startup.ServerID.String(), startup.StartupCommand)

	if err != nil {
		log.Println(err)
		return "", err
	}
	jobFile,err:=GenerateJobFile(startup.StartupCommand,startup.Variables)
	// // TODO:-using nomad client change the variables of the game server.
	// err = su.ChangeStartupVariables(startup.Variables, "cs2-server")
	// if err != nil {
	// 	su.logger.Error("error in updating variables", zap.Error(err))
	// 	return startup_id, err
	// }

		err=su.nomadClient.RegisterJob(ctx,jobFile)
		if err!=nil {
			log.Println(err)
			return "",err
		}

	return startup_id, nil

}

func (su *StartUpUsecase) GetStartup(ctx context.Context, id string) (*models.StartupInfo, error) {
	return su.repository.GetStartupParams(ctx, id)
}

func (su *StartUpUsecase) ChangeStartupVariables(variables map[string]interface{}, jobID string) error {
	// Create a Nomad API client
	return su.nomadClient.UpdateJobVariables(context.Background(), variables, jobID, jobID, jobID)

}

func (su *StartUpUsecase) DeleteStartupInfo(ctx context.Context, id string) error {
	return su.repository.DeleteStartupParams(ctx, id)
}

func (su *StartUpUsecase) GetGameEnvironments(ctx context.Context, game_name string) ([]string, error) {
	return su.repository.GetGameEnvironments(ctx, game_name)
}

func (su *StartUpUsecase) GetGameStartupCommand(ctx context.Context, serverID uuid.UUID) (string, error) {
	game, err := su.repository.GetGame(ctx, serverID)
	log.Println("game----------->", game)
	if err != nil {
		return "", err
	}
	startup_command, err := su.repository.GetStartupCommand(ctx, game)

	if err != nil {
		return "", err

	}
	return startup_command, nil
}

func (su *StartUpUsecase) GetGameInfo(ctx context.Context, game string) (*models.Game, error) {
	return su.repository.GetGameDetailedInfo(ctx, game)
}

func (su *StartUpUsecase) GetDefaultStartupCommand(ctx context.Context, game string) (string, error) {
	gameInfo, err := su.repository.GetGameDetailedInfo(ctx, game)
	if err != nil {
		return "", err
	}
	defaultCommand, err := generateDefaultStartupCommand(gameInfo.DefaultStartupCommand, gameInfo.DefaultVariables)
	if err != nil {
		return "", err
	}
	return defaultCommand, nil
}

// func (su *StartUpUsecase) RunTheScript(ctx context.Context, gameName string) {

// 	gameInfo, err := su.repository.GetGameDetailedInfo(ctx, gameName)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	fillHCLTemplate()

// }

// func(su *StartUpUsecase)GetServerStartupCommand(ctx context.Context,serverName string)(string,error){

// }

func generateStartupCommand(command string, variables map[string]interface{}) (string, error) {

	log.Println(command, variables)
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		command = strings.ReplaceAll(command, placeholder, fmt.Sprintf("%v", value))
	}

	return command, nil

}

func generateDefaultStartupCommand(command string, variables []string) (string, error) {
	log.Println("command:---", command)
	log.Println("varible:---", variables)

	for _, v := range variables {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return "", errors.New("invalid variable format")
		}
		key := parts[0]
		value := parts[1]

		// Remove surrounding quotes from value
		value = strings.Trim(value, "\"")

		// Replace all occurrences of key with value in the command string
		re := regexp.MustCompile("\\{\\{" + key + "\\}\\}")
		command = re.ReplaceAllString(command, value)
	}

	return command, nil
}

func GenerateJobFile(command string,variables map[string]interface{}) (string, error) {

	params := ServerParams{
		StartupCommand: command,
		Variables: variables,

	}

	// Job file template
	jobTemplate := `
job "minecraft-server" {
  datacenters = ["dc1"]

  group "minecraft-group" {
    task "minecraft-task" {
      driver = "raw_exec"

      config {
        command = "/bin/bash"
        args = [
          "-c",
          <<EOF
#!/bin/bash
SERVER_DIR="/mnt/server"
SERVER_JARFILE="$SERVER_DIR/{{index .Variables "SERVER_JAR"}}"
EULA_FILE="$SERVER_DIR/eula.txt"

# Agree to the EULA

# Update package repositories and install required packages
sudo apt update
sudo apt install -y curl jq

# Retrieve the latest version of Minecraft from Mojang's version_manifest.json
LATEST_VERSION=$(curl -sSL https://launchermeta.mojang.com/mc/game/version_manifest.json | jq -r '.latest.release')

# Retrieve the download URL for the server JAR file
MANIFEST_URL=$(curl -sSL https://launchermeta.mojang.com/mc/game/version_manifest.json | jq -r ".versions[] | select(.id == \"$LATEST_VERSION\") | .url")
DOWNLOAD_URL=$(curl -sSL $MANIFEST_URL | jq -r '.downloads.server.url')

# Download the server JAR file
curl -o "$SERVER_JARFILE" "$DOWNLOAD_URL"

# Notify the user that the download is complete
echo "Minecraft server JAR file downloaded successfully!"

# Run the provided startup command
java -Xms256M -Xmx512M -Dcom.mojang.eula.agree=true -jar "$SERVER_JARFILE" nogui

# Prompt the user for input
# Start the Minecraft server
# Insert your command here to start the Minecraft server
EOF
        ]
      }

      resources {
        cpu    = 500  # adjust according to your server requirements
        memory = 512  # adjust according to your server requirements
        network {
          mbits = 10
          port "minecraft" {}
        }
      }
    }
  }
}
`

	// Prepare template with job template
	tmpl, err := template.New("job").Parse(jobTemplate)
	if err != nil {
		log.Println("error in preparing job template",err)
		return "", err
	}

	// Create buffer to store filled template
	var filledTemplate strings.Builder

	// Execute template with job params
	err = tmpl.Execute(&filledTemplate, params)
	if err != nil {
		log.Println("error in executing job template",err)
		return "", err
	}

	// Convert filled template to string and return

	log.Println(filledTemplate.String())
	return filledTemplate.String(), nil

}
