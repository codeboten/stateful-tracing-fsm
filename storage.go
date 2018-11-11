package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/codeboten/fsm"

	"github.com/hashicorp/consul/api"
	"github.com/honeycombio/beeline-go/trace"
)

func getKV() *api.KV {
	// Get a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		fmt.Println("Failed to connect to consul KV")
		os.Exit(1)
	}

	// Get a handle to the KV API
	return client.KV()
}

func getContext(ctx context.Context) (context.Context, error) {
	pair, _, err := getKV().Get("Trace", nil)
	if err == nil && pair != nil {
		ctx, _ = trace.NewTrace(ctx, string(pair.Value))
		return ctx, nil
	}
	return ctx, errors.New("no previous context stored")
}

func loadState(fsm *fsm.FSM) {
	// Lookup the pair
	pair, _, err := getKV().Get("ApplicationState", nil)
	if err == nil && pair != nil {
		fsm.SetState(string(pair.Value))
	}
}

func storeState(fsm *fsm.FSM, span *trace.Span) {
	kv := getKV()
	// PUT a new KV pair
	p := &api.KVPair{Key: "ApplicationState", Value: []byte(fsm.Current())}
	_, err := kv.Put(p, nil)
	if err != nil {
		fmt.Println("failed to store ApplicationState")
		return
	}
	if span != nil {
		p = &api.KVPair{Key: "Trace", Value: []byte(span.SerializeHeaders())}
		_, err = kv.Put(p, nil)
		if err != nil {
			fmt.Println("failed to store Trace")
		}
	} else {
		kv.Delete("Trace", nil)
	}
}
