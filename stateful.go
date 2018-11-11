package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"

	"github.com/codeboten/fsm"
)

var (
	eventLaunchNode  = "launch-node"
	eventNodeUp      = "node-up"
	eventNodeHealthy = "node-healthy"
	eventRemoveNode  = "remove-old-node"
)

var (
	stateIdle       = "idle"
	stateLaunching  = "launching"
	stateValidating = "validating"
	stateRemoving   = "removing"
)

func fsmCallbackWrapper(ctx context.Context, f func(e *fsm.Event)) func(e *fsm.Event) {
	return func(e *fsm.Event) {
		_, span := beeline.StartSpan(ctx, e.Event)
		defer span.Send()
		f(e)
	}
}

func getStateMachine(ctx context.Context) *fsm.FSM {
	return fsm.NewFSM(
		stateIdle,
		fsm.Events{
			{Name: eventLaunchNode, Src: []string{stateIdle}, Dst: stateLaunching},
			{Name: eventNodeUp, Src: []string{stateLaunching}, Dst: stateValidating},
			{Name: eventNodeHealthy, Src: []string{stateValidating}, Dst: stateRemoving},
			{Name: eventRemoveNode, Src: []string{stateRemoving}, Dst: stateIdle},
		},
		fsm.Callbacks{
			eventLaunchNode: fsmCallbackWrapper(ctx, func(e *fsm.Event) {
				fmt.Println("launch node event: " + e.FSM.Current())
			}),
			eventNodeHealthy: fsmCallbackWrapper(ctx, func(e *fsm.Event) {
				fmt.Println("node healthy: " + e.FSM.Current())
			}),
			eventNodeUp: fsmCallbackWrapper(ctx, func(e *fsm.Event) {
				time.Sleep(time.Second * 5)
				fmt.Println("node up: " + e.FSM.Current())
			}),
			eventRemoveNode: fsmCallbackWrapper(ctx, func(e *fsm.Event) {
				fmt.Println("node removed: " + e.FSM.Current())
			}),
		},
	)
}

func run() {
	var span *trace.Span

	ctx, err := getContext(context.Background())
	if err != nil {
		ctx, span = beeline.StartSpan(ctx, "init")
	} else {
		ctx, span = beeline.StartSpan(ctx, "resume")
	}

	defer span.Send()

	fsm := getStateMachine(ctx)
	loadState(fsm)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		storeState(fsm, trace.GetTraceFromContext(ctx).GetRootSpan())
		span.Send()
		beeline.Flush(ctx)
		fmt.Println(sig)
		os.Exit(0)
	}()

	fmt.Println(fsm.Current())

	err = fsm.Event(eventLaunchNode)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("1:" + fsm.Current())

	err = fsm.Event(eventNodeUp)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("2:" + fsm.Current())

	err = fsm.Event(eventNodeHealthy)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("3:" + fsm.Current())

	err = fsm.Event(eventRemoveNode)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("4:" + fsm.Current())
	storeState(fsm, nil)
	defer beeline.Flush(ctx)
}

func main() {
	beeline.Init(beeline.Config{
		WriteKey:    os.Getenv("HONEYCOMB_KEY"),
		Dataset:     os.Getenv("HONEYCOMB_DATASET"),
		ServiceName: "node-manager",
	})
	run()
}
