package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const N_CONTAINERS = 8
const ALLOWED_CPU_P = 45
const NANO_CPUS = 500000000
const CONTAINER_PAUSE = 200
const WATCHER_INTERVAL = 200

func main() {

	// Holds the running containers
	rcs := ConcurrentSlice{}
	// Holds the paused containers
	pcs := ConcurrentSlice{}

	// Quit Pause channel - Used to force-quit the rotine that handles CPU usage
	qpc := make(chan bool)
	// Finished channel - Used when a container reaches the WaitConditionNotRunning status
	fc := make(chan string)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	go func() {
		// This go rotine is responsible for creating and starting N_CONTAINERS
		for i := 1; i <= N_CONTAINERS; i++ {
			resp, err := cli.ContainerCreate(ctx,
				&container.Config{
					Image: "progrium/stress",
					Cmd:   []string{"--cpu", "1", "--timeout", "30s"},
				},
				&container.HostConfig{
					Resources: container.Resources{
						NanoCPUs: NANO_CPUS,
					},
				},
				nil,
				nil,
				"")
			if err != nil {
				panic(err)
			}

			rcs.append(resp.ID)
			time.Sleep(CONTAINER_PAUSE * time.Millisecond)

			if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
				panic(err)
			}

			go func() {
				// This might be kind of awkward, to launch a go rotine inside other go-rotine's for-loop
				// the thing is: ContainerWait waits for one specific container
				// We need to ContainerWait each ContainerStart
				statusCh, errCh := cli.ContainerWait(
					ctx,
					resp.ID,
					container.WaitConditionNotRunning,
				)
				// Message main rotine
				select {
				case err := <-errCh:
					if err != nil {
						log.Fatal(err)
						fc <- resp.ID
					}
				case status := <-statusCh:
					log.Printf("status.StatusCode: %#+v\n", status.StatusCode)
					fc <- resp.ID
				}
			}()
		}
	}()

	go func() {
		// This go rotine will look after our CPU usage
		// whenever docker is under a heavy-load this rotine will take action and pause
		// the latest started container
		for {
			select {
			case <-qpc:
				return
			default:
				time.Sleep(WATCHER_INTERVAL * time.Millisecond)
				p := CPUPercentage()

				if p > ALLOWED_CPU_P {
					rc, err := rcs.remove()
					if err == nil {
						pcs.append(rc)
						cli.ContainerPause(ctx, rc.(string))
						log.Println("Paused", rc.(string))
					}
				}
				if p <= ALLOWED_CPU_P {
					rc, err := pcs.remove()
					if err == nil {
						rcs.append(rc)
						cli.ContainerUnpause(ctx, rc.(string))
						log.Println("Unpaused", rc.(string))
					}
				}
			}
		}
	}()

	// Wait for all the container waiting rotines to finish
	for i := 1; i <= N_CONTAINERS; i++ {
		fmt.Println(<-fc)
	}

	// Stop the CPU watcher rotine
	qpc <- true
}
