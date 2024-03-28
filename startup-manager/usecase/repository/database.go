package repository

import (
	"context"
	"encoding/json"
	"log"
	"startup-manager/core/models"
	core "startup-manager/core/postgres"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type StartupRepository struct {
	core.Postgres
}

func NewStartupRepository(db core.Postgres) *StartupRepository {
	return &StartupRepository{
		db,
	}
}

func (sr *StartupRepository) AddStartupParams(ctx context.Context, startup *models.StartupInfo) (string, error) {
	var starup_id string
	query := `INSERT INTO startups_info(server_id,variables,command)VALUES($1,$2,$3)returning id`
	startupVariablesJson, err := json.Marshal(startup.Variables)
	if err != nil {
		return "", err
	}
	log.Println(startupVariablesJson)
	err = sr.DB.QueryRowContext(ctx, query, startup.ServerID, startupVariablesJson, startup.StartupCommand).Scan(&starup_id)

	if err != nil {
		log.Println("err at 38", err)
		return "", err
	}
	return starup_id, nil
}

// Get startupInfo based on provided id
func (sr *StartupRepository) GetStartupParams(ctx context.Context, id string) (*models.StartupInfo, error) {
	query := `SELECT variables FROM startups_info WHERE id=$1`

	var variables interface{}

	err := sr.DB.Get(&variables, query, id)
	if err != nil {
		log.Println("error is", err)
		return nil, err
	}
	var variablesmap map[string]interface{}
	json.Unmarshal(variables.([]byte), &variablesmap)

	var startup_info_str models.StartupInfo
	startup_info_str.Variables = variablesmap
	query3 := "SELECT id,server_id,created_at,updated_at,deleted_at from startups_info WHERE id=$1"
	err = sr.DB.Get(&startup_info_str, query3, id)
	if err != nil {
		log.Println("err at 63", err)
		return nil, err
	}
	log.Println(startup_info_str)
	return &startup_info_str, nil
}

func (sr *StartupRepository) DeleteStartupParams(ctx context.Context, id string) error {
	query := "UPDATE startups_info SET deleted_at=now() WHERE id=$1"

	_, err := sr.DB.ExecContext(ctx, query, id)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (sr *StartupRepository) UpdateStartupParams(ctx context.Context, startup *models.StartupInfo) (string, error) {
	var startupID string
	startup_variables_json, err := json.Marshal(startup.Variables)
	if err != nil {
		log.Println(err)
		return "", err
	}
	err = sr.DB.QueryRowContext(ctx, "UPDATE startups_info SET variables=$1,updated_at=now() WHERE id=$2 RETURNING id", startup_variables_json, startup.ID).Scan(&startupID)
	if err != nil {
		return "", err
	}
	return startupID, nil
}

func (sr *StartupRepository) GetGameEnvironments(ctx context.Context, game_name string) ([]string, error) {
	query := "SELECT envs from games where name=$1"

	var envs pq.StringArray
	err := sr.DB.QueryRowContext(ctx, query, game_name).Scan(&envs)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return envs, nil

}

func (sr *StartupRepository) GetGame(ctx context.Context, serverId uuid.UUID) (string, error) {

	query := "SELECT game_name from gs_info WHERE id=$1"

	var game string
	err := sr.DB.QueryRowContext(ctx, query, serverId).Scan(&game)
	if err != nil {
		return "", nil
	}
	return game, nil
}

func (sr *StartupRepository) GetStartupCommand(ctx context.Context, game string) (string, error) {
	query := "SELECT default_startup_command from games WHERE name=$1"

	var default_startup_command string
	err := sr.DB.QueryRowContext(ctx, query, game).Scan(&default_startup_command)
	if err != nil {
		return "", nil
	}
	return default_startup_command, nil
}
func (sr *StartupRepository) GetGameDetailedInfo(ctx context.Context, game string) (*models.Game, error) {
	log.Println(game)
	query := "SELECT * FROM games WHERE name=$1"

	var gameDetail models.Game

	err := sr.DB.QueryRowContext(ctx, query, game).Scan(
		&gameDetail.ID,
		&gameDetail.Name,
		&gameDetail.Description,
		&gameDetail.Image,
		&gameDetail.Envs,
		&gameDetail.Ports,
		&gameDetail.Volumes,
		&gameDetail.CPU,
		&gameDetail.Memory,
		&gameDetail.Command,
		&gameDetail.Args,
		&gameDetail.DefaultStartupCommand,
		&gameDetail.DefaultVariables,
		&gameDetail.WithDB,
		&gameDetail.CreatedAt,
		&gameDetail.UpdatedAt,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &gameDetail, nil
}

func (sr *StartupRepository) GetServerStartupCommand(ctx context.Context, serverName string) (string, error) {
	log.Println(serverName)

	query := "SELECT command from gs_info WHERE serverName=$1"

	var startupCommand string

	err := sr.DB.QueryRowContext(ctx, query, serverName).Scan(&startupCommand)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return startupCommand, nil
}

func (sr *StartupRepository) UpdateGSCommand(ctx context.Context, serverID string, command string) error {
	_, err := sr.DB.ExecContext(ctx, "UPDATE gs_info SET command=$1, updated_at=now() WHERE id=$2", command, serverID)
	if err != nil {
		return err
	}

	return nil
}
