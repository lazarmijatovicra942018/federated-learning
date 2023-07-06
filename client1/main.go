package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"federated-learning/messages"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/automanaged"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
)

var sys *actor.ActorSystem = nil

type clientActor struct {
	system *actor.ActorSystem
}

type DTO struct {
	Layer1WeightsMatrix [][]float32 `json:"layer1_weights_matrix"`
	Bias1               []float32   `json:"bias1"`
	Layer2WeightsMatrix [][]float32 `json:"layer2_weights_matrix"`
	Bias2               []float32   `json:"bias2"`
	Layer3WeightsMatrix [][]float32 `json:"layer3_weights_matrix"`
	Bias3               []float32   `json:"bias3"`
}

func (p *clientActor) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case int:

		resp, err := http.Get("http://localhost:5000/get_weights")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println(resp.Body)

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var dto DTO
		err = json.Unmarshal(body, &dto)
		if err != nil {
			fmt.Printf("Failed to deserialize the response body: %s\n", err)
			return
		}

		var layer1WeightsMatrix []*messages.Row

		for i := 0; i < len(dto.Layer1WeightsMatrix); i++ {
			row := messages.Row{
				Values: dto.Layer1WeightsMatrix[i],
			}
			layer1WeightsMatrix = append(layer1WeightsMatrix, &row)
		}

		var layer2WeightsMatrix []*messages.Row

		for i := 0; i < len(dto.Layer2WeightsMatrix); i++ {
			row := messages.Row{
				Values: dto.Layer2WeightsMatrix[i],
			}
			layer2WeightsMatrix = append(layer2WeightsMatrix, &row)
		}

		var layer3WeightsMatrix []*messages.Row

		for i := 0; i < len(dto.Layer3WeightsMatrix); i++ {
			row := messages.Row{
				Values: dto.Layer3WeightsMatrix[i],
			}
			layer3WeightsMatrix = append(layer3WeightsMatrix, &row)
		}

		mess := &messages.DTO{
			Layer1WeightsMatrix: layer1WeightsMatrix,
			Bias1:               dto.Bias1,
			Layer2WeightsMatrix: layer2WeightsMatrix,
			Bias2:               dto.Bias2,
			Layer3WeightsMatrix: layer3WeightsMatrix,
			Bias3:               dto.Bias3,
		}

		coordinatorPid := cluster.GetCluster(p.system).Get("client-1", "CoordinatorCluster")

		future := ctx.RequestFuture(coordinatorPid, mess, 180*time.Second)
		result, err := future.Result()
		if err != nil {
			log.Print(err.Error())
			return
		}
		log.Printf("Received %v", result)

	}

}

func main() {
	// Set up actor system
	sys = actor.NewActorSystem()

	// Prepare a remote env that listens to 8081
	//config := remote.Configure("127.0.0.1", 8081)
	config := remote.Configure("192.168.0.113", 8081)

	// Configure a cluster on top of the above remote env
	clusterProvider := automanaged.NewWithConfig(1*time.Second, 6330, "192.168.0.113:6331")
	lookup := disthash.New()
	clusterConfig := cluster.Configure("cluster-coordinator", clusterProvider, lookup, config)
	c := cluster.New(sys, clusterConfig)

	// Manage the cluster client's lifecycle
	c.StartClient() // Configure as a client

	// Start a ping actor that periodically sends a "ping" payload to the "Ponger" cluster grain
	clientProps := actor.PropsFromProducer(func() actor.Actor {
		return &clientActor{
			system: sys,
		}
	})

	clientPid := sys.Root.Spawn(clientProps)

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, os.Interrupt, os.Kill)

	sys.Root.Send(clientPid, 0)

	for {
		select {

		case <-finish:
			log.Print("Finish")
			return

		}
	}
}
