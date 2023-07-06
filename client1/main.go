package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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
	Layer1WeightsMatrix [][]float64 `json:"layer1_weights_matrix"`
	Bias1               []float64   `json:"bias1"`
	Layer2WeightsMatrix [][]float64 `json:"layer2_weights_matrix"`
	Bias2               []float64   `json:"bias2"`
	Layer3WeightsMatrix [][]float64 `json:"layer3_weights_matrix"`
	Bias3               []float64   `json:"bias3"`
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

		// Now you can access the DTO object in Go

		//layer1WeightsMatrix := dto.Layer1WeightsMatrix
		//bias1 := dto.Bias1
		//	layer2WeightsMatrix := dto.Layer2WeightsMatrix
		//	bias2 := dto.Bias2
		//	layer3WeightsMatrix := dto.Layer3WeightsMatrix
		//bias3 := dto.Bias3

		// Now you can work with each field as needed
		//fmt.Println("Layer 1 Weights Matrix:", layer1WeightsMatrix)
		//fmt.Println("Bias 1:", bias1)
		/*	fmt.Println("Layer 2 Weights Matrix:", layer2WeightsMatrix)
			fmt.Println("Bias 2:", bias2)
			fmt.Println("Layer 3 Weights Matrix:", layer3WeightsMatrix)
		*/
		//fmt.Println("Bias 3:", bias3)

		// Print the response

		//variableType := reflect.TypeOf(body)

		//fmt.Println(variableType)

		//p.weights = body
		/*
			var w []float64

			err = json.Unmarshal(string(body), &w)
			if err != nil {
					fmt.Println("\n\n\n\n\n\nError:", err)
					return
			}
				// Print the weights
			fmt.Println(w)
		*/

		mess := &messages.DTO{
			layer1_weights_matrix: dto.Layer1WeightsMatrix,
			bias1:                 dto.Bias1,
			layer2_weights_matrix: dto.Layer2WeightsMatrix,
			bias2:                 dto.Bias2,
			layer3_weights_matrix: dto.Layer3WeightsMatrix,
			bias3:                 dto.Bias3,
		}

		/*

				grainPid := cluster.GetCluster(p.system).Get("ponger-1", "Ponger")
				log.Print(grainPid)
				log.Print("\n\n\n\n\n\n\n\n\n")

				future := ctx.RequestFuture(grainPid, ping, 3*time.Second)
				result, err := future.Result()
				if err != nil {
					log.Print(err.Error())
					return
				}
				log.Printf("Received %v", result)

			case *messages.PongMessage:
				// Never comes here.
				// When the pong actor responds to the sender,
				// the sender is not a ping actor but a future process.
				log.Print("Received pong message")
		*/
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
