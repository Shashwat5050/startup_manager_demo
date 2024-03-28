package nomadapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	nomadApi "github.com/hashicorp/nomad/api"
)

const (
	// Stdout is the stdoutLogType stream
	stdoutLogType = "stdout"
	// Stderr is the stderrLogType stream
	stderrLogType = "stderr"
)

type NomadClient struct {
	client *nomadApi.Client
}

func NewNomadClient(nomadURL string) (*NomadClient, error) {
	nc, err := nomadApi.NewClient(&nomadApi.Config{
		Address:   nomadURL,
		TLSConfig: &nomadApi.TLSConfig{Insecure: true},
	})
	if err != nil {
		return nil, err
	}

	// check if client is working
	_, _, err = nc.Jobs().List(nil)
	if err != nil {
		return nil, err
	}

	return &NomadClient{
		client: nc,
	}, nil
}

func (n *NomadClient) RegisterJob(ctx context.Context, jobHCL string) error {
	job, err := n.client.Jobs().ParseHCL(jobHCL, true)
	if err != nil {
		return fmt.Errorf("could not parse job hcl: %w", err)
	}

	_, err = n.client.Namespaces().Register(&nomadApi.Namespace{Name: *job.Namespace}, &nomadApi.WriteOptions{})
	if err != nil {
		return fmt.Errorf("could not register namespace: %w", err)
	}

	_, _, err = n.client.Jobs().Register(job, &nomadApi.WriteOptions{Namespace: *job.Namespace})
	if err != nil {
		return fmt.Errorf("could not register job: %w", err)
	}

	return nil
}

func (n *NomadClient) CheckJobStatus(ctx context.Context, jobID, namespace string) (string, error) {
	allocs, err := n.getAllocations(ctx, jobID, namespace)
	if err != nil {
		return "", err
	}

	alloc, _, err := n.client.Allocations().Info(allocs[0].ID, &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return "", err
	}

	return alloc.ClientStatus, nil
}

// Write function that starts job and allocation with given ID
func (n *NomadClient) StartJob(ctx context.Context, jobID string) error {
	job, _, err := n.client.Jobs().Info(jobID, &nomadApi.QueryOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	stop := false
	job.Stop = &stop

	_, _, err = n.client.Jobs().RegisterOpts(job, &nomadApi.RegisterOptions{
		EnforceIndex: true,
		ModifyIndex:  *job.JobModifyIndex,
	}, &nomadApi.WriteOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	return nil
}

// StopJob stops the job and all allocations
func (n *NomadClient) StopJob(ctx context.Context, jobID string) error {
	_, _, err := n.client.Jobs().Info(jobID, &nomadApi.QueryOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	_, _, err = n.client.Jobs().Deregister(jobID, false, &nomadApi.WriteOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	return nil
}

func (n *NomadClient) RestartJob(ctx context.Context, jobID string) error {
	_, _, err := n.client.Jobs().Info(jobID, &nomadApi.QueryOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	allocs, err := n.getAllocations(ctx, jobID, jobID)
	if err != nil {
		return err
	}

	alloc, _, err := n.client.Allocations().Info(allocs[0].ID, &nomadApi.QueryOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	err = n.client.Allocations().Restart(alloc, jobID, &nomadApi.QueryOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	return nil
}

func (n *NomadClient) DeleteJob(ctx context.Context, jobID string) error {
	_, _, err := n.client.Jobs().Info(jobID, &nomadApi.QueryOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	_, _, err = n.client.Jobs().Deregister(jobID, true, &nomadApi.WriteOptions{Namespace: jobID})
	if err != nil {
		return err
	}

	return nil
}

func (n *NomadClient) RunCommand(ctx context.Context, jobID, namespace string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) (int, error) {
	allocs, err := n.getAllocations(ctx, jobID, namespace)
	if err != nil {
		return 0, err
	}

	alloc, _, err := n.client.Allocations().Info(allocs[0].ID, &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return 0, err
	}

	var termSizeCh chan nomadApi.TerminalSize

	command := []string{cmd}
	command = append(command, args...)

	exitCode, err := n.client.Allocations().Exec(ctx, alloc, jobID,
		true, command, stdin, stdout, stderr, termSizeCh, &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return 0, err
	}

	return exitCode, err
}

func (n *NomadClient) GetAllocationNetwork(ctx context.Context, jobID, namespace string) (string, []int, error) {
	allocs, err := n.getAllocations(ctx, jobID, namespace)
	if err != nil {
		return "", nil, err
	}

	alloc, _, err := n.client.Allocations().Info(allocs[0].ID, &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return "", nil, err
	}

	if alloc.Resources == nil || len(alloc.Resources.Networks) == 0 {
		return "", nil, errors.New("no network resources")
	}

	ip := alloc.Resources.Networks[0].IP

	var ports []int
	for _, port := range alloc.Resources.Networks[0].DynamicPorts {
		if port.Value == 0 {
			continue
		}
		ports = append(ports, port.Value)
	}

	return ip, ports, nil
}

func (n *NomadClient) GetNodeIP(ctx context.Context, gsID string) (string, error) {
	allocs, err := n.getAllocations(ctx, gsID, gsID)
	if err != nil {
		return "", err
	}

	alloc, _, err := n.client.Allocations().Info(allocs[0].ID, &nomadApi.QueryOptions{Namespace: gsID})
	if err != nil {
		return "", err
	}

	if alloc.NodeID == "" {
		return "", errors.New("no node id")
	}

	node, _, err := n.client.Nodes().Info(alloc.NodeID, &nomadApi.QueryOptions{Namespace: gsID})
	if err != nil {
		return "", err
	}

	addr := node.HTTPAddr
	if addr == "" {
		return "", errors.New("no node address")
	}

	// obtain IP from addr
	ip := strings.Split(addr, ":")[0]
	if ip == "" {
		return "", errors.New("no node ip")
	}

	return ip, nil
}

func (n *NomadClient) GetStats(ctx context.Context, id, namespace string) ([]byte, error) {
	allocs, err := n.getAllocations(ctx, id, namespace)
	if err != nil {
		return nil, err
	}

	stats, err := n.client.Allocations().Stats(&nomadApi.Allocation{ID: allocs[0].ID}, &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}

	return json.Marshal(stats)
}

func (n *NomadClient) GetLogs(ctx context.Context, id, namespace, stdType string, offset int64) ([]byte, error) {
	allocs, err := n.getAllocations(ctx, id, namespace)
	if err != nil {
		return nil, err
	}

	alloc, _, err := n.client.Allocations().Info(allocs[0].ID, &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}

	var logType string
	switch stdType {
	case stdoutLogType:
		logType = nomadApi.FSLogNameStdout
	case stderrLogType:
		logType = nomadApi.FSLogNameStderr
	default:
		return nil, errors.New("invalid std type")
	}

	logCh, errCh := n.client.AllocFS().Logs(alloc, false, id, logType, nomadApi.OriginStart, int64(offset), ctx.Done(), &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}

	select {
	case log := <-logCh:
		return log.Data, nil
	case err := <-errCh:
		return nil, err
	}
}

func (n *NomadClient) GetSftpPort(ctx context.Context, jobID, namespace string) (int, error) {
	allocs, err := n.getAllocations(ctx, jobID, namespace)
	if err != nil {
		return -1, err
	}

	alloc, _, err := n.client.Allocations().Info(allocs[0].ID, &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return -1, err
	}

	if alloc.Resources == nil || len(alloc.Resources.Networks) == 0 {
		return -1, errors.New("no network resources")
	}

	sftpPort := -1
	for _, port := range alloc.Resources.Networks[0].DynamicPorts {
		if port.Label == "port-sftp" {
			sftpPort = port.Value
			break
		}
	}

	if sftpPort == -1 {
		return -1, errors.New("no sftp port found")
	}

	return sftpPort, nil
}

func (n *NomadClient) Name() string {
	return "nomad"
}

func (n *NomadClient) CheckHealth(ctx context.Context) error {
	_, _, err := n.client.Jobs().List(nil)
	if err != nil {
		return err
	}

	return nil
}

func (n *NomadClient) CheckReadiness(ctx context.Context) error {
	return n.CheckHealth(ctx)
}

func (n *NomadClient) getAllocations(ctx context.Context, jobID, namespace string) ([]*nomadApi.AllocationListStub, error) {
	allocs, _, err := n.client.Jobs().Allocations(jobID, false, &nomadApi.QueryOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}

	if len(allocs) == 0 {
		return nil, errors.New("no allocations")
	}

	sort.Slice(allocs, func(i, j int) bool {
		return allocs[i].CreateTime > allocs[j].CreateTime
	})

	return allocs, nil
}

func (n *NomadClient) UpdateJobVariables(ctx context.Context, variables map[string]interface{}, jobID, namespace string, serverName string) error {

	log.Println("jobID is ", jobID)
	job, _, err := n.client.Jobs().Info(jobID,&nomadApi.QueryOptions{})
	if err != nil {
		return err
	}
	log.Println(variables)
	for _, taskGroup := range job.TaskGroups {
		for _, task := range taskGroup.Tasks {
			if task.Name == serverName {
				// Create a new map for environment variables
				newEnv := make(map[string]string)

				// Copy existing environment variables to the new map
				for key, value := range task.Env {
					newEnv[key] = value
				}

				// Update the environment variable or add a new one
				for key, value := range variables {
					log.Println(key,value,"key,val")
					newEnv[key] = value.(string)
				}

				// Set the updated environment variables for the task
				task.Env = newEnv

			}
		}
	}

	// Submit the updated job specification

	writeOptions := &nomadApi.WriteOptions{}
	_, _, err = n.client.Jobs().Register(job, writeOptions)
	if err != nil {
		return err
	}

	return nil

}
