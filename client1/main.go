package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/automanaged"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
)

var cnt uint64 = 0

var sys *actor.ActorSystem = nil

type clientActor struct {
	system  *actor.ActorSystem
	weights []uint8
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

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Print the response

		variableType := reflect.TypeOf(body)

		fmt.Println(variableType)

		var w []uint8
		err = json.Unmarshal(body, &w)
		if err != nil {
			fmt.Println("\n\n\n\n\n\nError:", err)
			return
		}
		// Print the weights
		fmt.Println(w)
		/*
			mess := &messages.ClientMessage{
					weights : body

			}*/

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
			system:  sys,
			weights: nil,
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
