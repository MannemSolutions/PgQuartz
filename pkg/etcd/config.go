package etcd

import (
	"context"
	"time"
)

type Config struct {
	Endpoints   []string `yaml:"endpoints"`
	LockKey     string   `yaml:"lockKey"`
	LockTimeout string   `yaml:"lockTimeout"`
}

func (ec *Config) SetDefaults() {
	// Create a etcd client
	if len(ec.Endpoints) == 0 {
		ec.Endpoints = []string{"localhost:2379"}
	}
	if ec.LockTimeout == "" {
		ec.LockTimeout = "100h"
	}
}

func (ec Config) GetTimeoutDuration(parentContext context.Context) (context.Context, context.CancelFunc) {
	if ec.LockTimeout == "" {
		return parentContext, nil
	}
	lockDuration, err := time.ParseDuration(ec.LockTimeout)
	if err != nil {
		log.Fatal(err)
	}
	return context.WithTimeout(parentContext, lockDuration)
}
