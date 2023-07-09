package main

import (
	"encoding/csv"
	"federated-learning/messages"
	"fmt"
	"log"
	"os"
	"time"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/automanaged"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
)

var sys *actor.ActorSystem = nil

// var ip_addr_E = "192.168.1.9"
var ip_addr_E = "192.168.0.113"

type (
	Initializer struct {
		pid            *actor.PID
		coordinatorPID *actor.PID
		loggerPID      *actor.PID
		aggregatorPID  *actor.PID
	}
	Coordinator struct {
		pid           *actor.PID
		parentPID     *actor.PID
		loggerPID     *actor.PID
		aggregatorPID *actor.PID
		cluster       *cluster.Cluster
		chIn          chan DTO
		chOut         chan DTO
		finalWeights  DTO
	}
	Logger struct {
		pid            *actor.PID
		parentPID      *actor.PID
		aggregatorPID  *actor.PID
		coordinatorPID *actor.PID
		writer         csv.Writer
	}
	Aggregator struct {
		pid            *actor.PID
		parentPID      *actor.PID
		loggerPID      *actor.PID
		coordinatorPID *actor.PID
		chIn           chan DTO
		chOut          chan DTO
		clientWeights  []DTO
	}
	pidsDtos struct {
		initPID        *actor.PID
		coordinatorPID *actor.PID
		loggerPID      *actor.PID
		aggregatorPID  *actor.PID
	}
	DTO struct {
		Layer1WeightsMatrix [][]float32 `json:"layer1_weights_matrix"`
		Bias1               []float32   `json:"bias1"`
		Layer2WeightsMatrix [][]float32 `json:"layer2_weights_matrix"`
		Bias2               []float32   `json:"bias2"`
		Layer3WeightsMatrix [][]float32 `json:"layer3_weights_matrix"`
		Bias3               []float32   `json:"bias3"`
	}
	chDtos struct {
		chIn  chan DTO
		chOut chan DTO
	}
	Message struct {
		Content  string
		DateTime time.Time
		Role     string
	}
)

func (state *Initializer) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.PID:
		coorditatorProps := actor.PropsFromProducer(newCoordinatorActor)
		coorditatorPID := context.Spawn(coorditatorProps)
		aggregatorProps := actor.PropsFromProducer(newAggregatorActor)
		aggregatorPID := context.Spawn(aggregatorProps)
		loggerProps := actor.PropsFromProducer(newLoggerActor)
		loggerPID := context.Spawn(loggerProps)

		state.pid = msg
		state.aggregatorPID = aggregatorPID
		state.loggerPID = loggerPID
		state.coordinatorPID = coorditatorPID

		context.Send(loggerPID, pidsDtos{initPID: state.pid, loggerPID: loggerPID, aggregatorPID: aggregatorPID, coordinatorPID: coorditatorPID})
		context.Send(coorditatorPID, pidsDtos{initPID: state.pid, coordinatorPID: coorditatorPID, loggerPID: loggerPID, aggregatorPID: aggregatorPID})
		context.Send(aggregatorPID, pidsDtos{initPID: state.pid, aggregatorPID: aggregatorPID, loggerPID: loggerPID, coordinatorPID: coorditatorPID})
		context.Send(loggerPID, Message{Content: "All actors are created !", DateTime: time.Now(), Role: "Initializer"})
	}
}

func newInitializatorActor() actor.Actor {
	return &Initializer{}
}

func newCoordinatorActor() actor.Actor {
	return &Coordinator{}
}

func newAggregatorActor() actor.Actor {
	return &Aggregator{}
}

func newLoggerActor() actor.Actor {
	return &Logger{}
}

// change ip address of computer and turn of firewall
func (state *Coordinator) clusterSetup(context actor.Context) *cluster.Cluster {
	config := remote.Configure(ip_addr_E, 8080)
	provider := automanaged.NewWithConfig(1*time.Second, 6331, ip_addr_E+":6331")
	lookup := disthash.New()
	clusterKind := cluster.NewKind(
		"CoordinatorCluster",
		actor.PropsFromProducer(func() actor.Actor {
			return state
		}))
	clusterConfig := cluster.Configure("cluster-coordinator", provider, lookup, config, cluster.WithKinds(clusterKind))
	c := cluster.New(sys, clusterConfig)
	state.cluster = c

	return c
}

func (state *Aggregator) Funnel(in <-chan DTO, out chan<- DTO, context actor.Context) {
	for {
		var data DTO
		data, _ = <-in
		state.clientWeights = append(state.clientWeights, data)
		if len(state.clientWeights) == 2 {
			//proveri da li su dva u nizu, ako jesu uradi sta treba, stavi u out i return
			fmt.Println("obradjuje se fed avg")
			context.Send(state.loggerPID, Message{Content: "Proccesing weights !", DateTime: time.Now(), Role: "Aggregator"})
			dto1 := state.clientWeights[0]
			dto2 := state.clientWeights[1]

			var layer1WeightsMatrix [][]float32
			for i := 0; i < len(dto1.Layer1WeightsMatrix); i++ {
				var rowT []float32
				for j := 0; j < len(dto1.Layer1WeightsMatrix[i]); j++ {
					rowT = append(rowT, (dto1.Layer1WeightsMatrix[i][j]+dto2.Layer1WeightsMatrix[i][j])/2)

				}
				layer1WeightsMatrix = append(layer1WeightsMatrix, rowT)
			}

			var layer2WeightsMatrix [][]float32
			for i := 0; i < len(dto1.Layer2WeightsMatrix); i++ {
				var rowT []float32
				for j := 0; j < len(dto1.Layer2WeightsMatrix[i]); j++ {
					rowT = append(rowT, (dto1.Layer2WeightsMatrix[i][j]+dto2.Layer2WeightsMatrix[i][j])/2)

				}
				layer2WeightsMatrix = append(layer2WeightsMatrix, rowT)
			}

			var layer3WeightsMatrix [][]float32
			for i := 0; i < len(dto1.Layer3WeightsMatrix); i++ {
				var rowT []float32
				for j := 0; j < len(dto1.Layer3WeightsMatrix[i]); j++ {
					rowT = append(rowT, (dto1.Layer3WeightsMatrix[i][j]+dto2.Layer3WeightsMatrix[i][j])/2)

				}
				layer3WeightsMatrix = append(layer3WeightsMatrix, rowT)
			}

			var bias1 []float32
			for i := 0; i < len(dto1.Bias1); i++ {
				bias1 = append(bias1, (dto1.Bias1[i]+dto2.Bias1[i])/2)
			}

			var bias2 []float32
			for i := 0; i < len(dto1.Bias2); i++ {
				bias2 = append(bias2, (dto1.Bias2[i]+dto2.Bias2[i])/2)
			}

			var bias3 []float32
			for i := 0; i < len(dto1.Bias3); i++ {
				bias3 = append(bias3, (dto1.Bias3[i]+dto2.Bias3[i])/2)
			}

			finalDTO := DTO{
				Bias1:               bias1,
				Bias2:               bias2,
				Bias3:               bias3,
				Layer1WeightsMatrix: layer1WeightsMatrix,
				Layer2WeightsMatrix: layer2WeightsMatrix,
				Layer3WeightsMatrix: layer3WeightsMatrix,
			}

			context.Send(state.loggerPID, Message{Content: "Weights are agregated !", DateTime: time.Now(), Role: "Aggregator"})
			out <- finalDTO
			out <- finalDTO
			return
		}

	}
}

func (state *Coordinator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Restart:
		time.Sleep(3 * time.Second)
		state.cluster = state.clusterSetup(context)
		state.cluster.StartMember()
	case pidsDtos:
		if msg.initPID == nil {
		}
		state.parentPID = msg.initPID
		state.pid = msg.coordinatorPID
		state.loggerPID = msg.loggerPID
		state.aggregatorPID = msg.aggregatorPID
		state.chIn = make(chan DTO)
		state.chOut = make(chan DTO)

		state.cluster = state.clusterSetup(context)
		state.cluster.StartMember()
		context.Send(state.loggerPID, Message{Content: "Prowider is set !", DateTime: time.Now(), Role: "Coordinator"})

		context.Send(state.aggregatorPID, chDtos{chIn: state.chIn, chOut: state.chOut})

	case *messages.DTO:
		fmt.Println("Received client message")
		context.Send(state.loggerPID, Message{Content: "Received client message !", DateTime: time.Now(), Role: "Coordinator"})

		var layer1WeightsMatrixT [][]float32
		for i := 0; i < len(msg.Layer1WeightsMatrix); i++ {
			var rowT []float32
			row := msg.Layer1WeightsMatrix[i]
			for _, value := range row.Values {
				floatValue := float32(value)
				rowT = append(rowT, floatValue)
			}
			layer1WeightsMatrixT = append(layer1WeightsMatrixT, rowT)
		}

		var layer2WeightsMatrixT [][]float32
		for i := 0; i < len(msg.Layer2WeightsMatrix); i++ {
			var rowT []float32
			row := msg.Layer2WeightsMatrix[i]
			for _, value := range row.Values {
				floatValue := float32(value)
				rowT = append(rowT, floatValue)
			}
			layer2WeightsMatrixT = append(layer2WeightsMatrixT, rowT)
		}

		var layer3WeightsMatrixT [][]float32
		for i := 0; i < len(msg.Layer3WeightsMatrix); i++ {
			var rowT []float32
			row := msg.Layer3WeightsMatrix[i]
			for _, value := range row.Values {
				floatValue := float32(value)
				rowT = append(rowT, floatValue)
			}
			layer3WeightsMatrixT = append(layer3WeightsMatrixT, rowT)
		}

		dto := DTO{
			Layer1WeightsMatrix: layer1WeightsMatrixT,
			Bias1:               msg.Bias1,
			Layer2WeightsMatrix: layer2WeightsMatrixT,
			Bias2:               msg.Bias2,
			Layer3WeightsMatrix: layer3WeightsMatrixT,
			Bias3:               msg.Bias3,
		}

		state.chIn <- dto

		context.Send(state.loggerPID, Message{Content: "Client weights sent to agregator !", DateTime: time.Now(), Role: "Coordinator"})

		if len(state.finalWeights.Bias1) == 0 {
			var data DTO
			data, _ = <-state.chOut
			state.finalWeights = data
			fmt.Println("Dobijanje final tezina")
		}

		var layer1WeightsMatrixR []*messages.Row

		for i := 0; i < len(state.finalWeights.Layer1WeightsMatrix); i++ {
			row := messages.Row{
				Values: state.finalWeights.Layer1WeightsMatrix[i],
			}
			layer1WeightsMatrixR = append(layer1WeightsMatrixR, &row)
		}

		var layer2WeightsMatrixR []*messages.Row

		for i := 0; i < len(state.finalWeights.Layer2WeightsMatrix); i++ {
			row := messages.Row{
				Values: state.finalWeights.Layer2WeightsMatrix[i],
			}
			layer2WeightsMatrixR = append(layer2WeightsMatrixR, &row)
		}

		var layer3WeightsMatrixR []*messages.Row
		for i := 0; i < len(state.finalWeights.Layer3WeightsMatrix); i++ {
			row := messages.Row{
				Values: state.finalWeights.Layer3WeightsMatrix[i],
			}
			layer3WeightsMatrixR = append(layer3WeightsMatrixR, &row)
		}

		response := &messages.DTO{
			Layer1WeightsMatrix: layer1WeightsMatrixR,
			Bias1:               state.finalWeights.Bias1,
			Layer2WeightsMatrix: layer2WeightsMatrixR,
			Bias2:               state.finalWeights.Bias2,
			Layer3WeightsMatrix: layer3WeightsMatrixR,
			Bias3:               state.finalWeights.Bias3,
		}
		context.Respond(response)
		context.Send(state.loggerPID, Message{Content: "Sent agregated weights to client !", DateTime: time.Now(), Role: "Coordinator"})

	}
}

func (state *Logger) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case pidsDtos:
		if msg.initPID == nil {
		}
		state.parentPID = msg.initPID
		state.pid = msg.loggerPID
		state.aggregatorPID = msg.aggregatorPID
		state.coordinatorPID = msg.coordinatorPID

		/*
			//creating CSV file
			file, err := os.Create("loger.csv")
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			writer := csv.NewWriter(file)
			defer writer.Flush()

			header := []string{"DateTime", "Role", "Content"}
			err = writer.Write(header)
			if err != nil {
				log.Fatal(err)
			}
			state.writer = *writer
		*/
	case Message:
		fmt.Println("\nSecam se kao da gledam baka me u crkvu vodi kraj mene ljudi se ljube i vicu hristos se rodi \n\n")
		file, err := os.OpenFile("loger.csv", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// Create a CSV writer
		writer := csv.NewWriter(file)
		//	defer writer.Flush()

		message := []string{msg.DateTime.Format("2006-01-02 15:04:05"), msg.Role, msg.Content}
		err = writer.Write(message)
		if err != nil {
			log.Fatal(err)
		}
		defer writer.Flush()
	}
}

func (state *Aggregator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case pidsDtos:
		if msg.initPID == nil {
		}
		state.parentPID = msg.initPID
		state.pid = msg.aggregatorPID
		state.loggerPID = msg.loggerPID
		state.coordinatorPID = msg.coordinatorPID
	case chDtos:
		state.chIn = msg.chIn
		state.chOut = msg.chOut
		go state.Funnel(msg.chIn, msg.chOut, context)
	}
}

func main() {
	sys = actor.NewActorSystem()
	decider := func(reason interface{}) actor.Directive {
		fmt.Println("handling failure for child")
		return actor.RestartDirective
	}

	supervisor := actor.NewOneForOneStrategy(10, 1000, decider)
	rootContext := sys.Root
	props := actor.
		PropsFromProducer(newInitializatorActor,
			actor.WithSupervisor(supervisor))
	pid := rootContext.Spawn(props)

	rootContext.Send(pid, pid)

	_, _ = console.ReadLine()
}
