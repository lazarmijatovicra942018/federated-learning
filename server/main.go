package main

import (
	"federated-learning/messages"
	"fmt"
	"time"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/automanaged"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
)

var sys *actor.ActorSystem = nil
var ip_addr_E = "192.168.1.5"
var ip_addr_L = "192.168.0.113"

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
	}
	Logger struct {
		pid            *actor.PID
		parentPID      *actor.PID
		aggregatorPID  *actor.PID
		coordinatorPID *actor.PID
	}
	Aggregator struct {
		pid            *actor.PID
		parentPID      *actor.PID
		loggerPID      *actor.PID
		coordinatorPID *actor.PID
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

		state.cluster = state.clusterSetup(context)
		state.cluster.StartMember()
	case *messages.DTO:
		fmt.Println("Received client message")

		/*var layer1WeightsMatrixT [][]float32
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

		//koristi kanale u agregatoru i kordinatoru
		context.Send(state.aggregatorPID, dto)
		*/
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
