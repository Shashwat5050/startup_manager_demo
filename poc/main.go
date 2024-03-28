package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/nomad/api"
)

func main() {
	// Create a Nomad API client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	// Define the job ID
	jobID := "minecraft-server"

	// Retrieve the current job specification
	job, _, err := client.Jobs().Info(jobID, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Modify the environment variables for a specific task in a task group
	for _, taskGroup := range job.TaskGroups {
		for _, task := range taskGroup.Tasks {
			if task.Name == "minecraft-server" {
				// Create a new map for environment variables
				newEnv := make(map[string]string)

				// Copy existing environment variables to the new map
				for key, value := range task.Env {
					newEnv[key] = value
				}

				// Update the environment variable or add a new one
				newEnv["TYPE"] = "VANILLA"

				// Set the updated environment variables for the task
				task.Env = newEnv
			}
		}
	}

	// Submit the updated job specification

	writeOptions := &api.WriteOptions{}
	_, _, err = client.Jobs().Register(job, writeOptions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Job updated successfully!")
}
