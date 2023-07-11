package main

import (
	"bytes"
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

// change ip address of computer and turn of firewall
var ip_addr_E = "192.168.1.5"
var ip_addr_of_provider = "192.168.1.5"

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
		fmt.Println("Received update weights")
		if result == nil {
		}

		if result == nil {
		}
		resultDTO, ok := result.(*messages.DTO)
		if !ok {
			log.Print("Received unexpected result type")
			return
		}

		dto = DTO{}

		dto.Bias1 = resultDTO.Bias1
		for i := 0; i < len(resultDTO.GetLayer1WeightsMatrix()); i++ {
			row := resultDTO.GetLayer1WeightsMatrix()[i].GetValues()
			dto.Layer1WeightsMatrix = append(dto.Layer1WeightsMatrix, row)

		}

		dto.Bias2 = resultDTO.Bias2
		for i := 0; i < len(resultDTO.GetLayer2WeightsMatrix()); i++ {
			row := resultDTO.GetLayer2WeightsMatrix()[i].GetValues()
			dto.Layer2WeightsMatrix = append(dto.Layer2WeightsMatrix, row)

		}

		dto.Bias3 = resultDTO.Bias3
		for i := 0; i < len(resultDTO.GetLayer3WeightsMatrix()); i++ {
			row := resultDTO.GetLayer3WeightsMatrix()[i].GetValues()
			dto.Layer3WeightsMatrix = append(dto.Layer3WeightsMatrix, row)

		}

		setWeights("set_weights", dto)

	}

}

func setWeights(url string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON data: %s", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:5000/"+url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("Error:", err)

	} else {
		response := string(string(body))
		fmt.Println(response)
		fmt.Println(resp.Status)
	}

	return nil
}

func main() {
	sys = actor.NewActorSystem()
	config := remote.Configure(ip_addr_E, 8081)

	clusterProvider := automanaged.NewWithConfig(1*time.Second, 6330, ip_addr_of_provider+":6331")
	lookup := disthash.New()
	clusterConfig := cluster.Configure("cluster-coordinator", clusterProvider, lookup, config)
	c := cluster.New(sys, clusterConfig)

	c.StartClient()

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
