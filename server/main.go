package main

import (
	"fmt"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
)

var sys *actor.ActorSystem = nil

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

		context.Send(loggerPID, pidsDtos{initPID: state.pid, loggerPID: loggerPID})
		context.Send(coorditatorPID, pidsDtos{initPID: state.pid, coordinatorPID: coorditatorPID, loggerPID: loggerPID, aggregatorPID: aggregatorPID})
		context.Send(aggregatorPID, pidsDtos{initPID: state.pid, aggregatorPID: aggregatorPID})

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

func (state *Coordinator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case pidsDtos:
		if msg.initPID == nil {
		}
		state.parentPID = msg.initPID
		state.pid = msg.coordinatorPID
		state.loggerPID = msg.loggerPID
		state.aggregatorPID = msg.aggregatorPID

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
	sys := actor.NewActorSystem()
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
