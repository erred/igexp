// Copyright 2018 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package function

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"cloud.google.com/go/logging"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

func init() {
	config = &configuration{}
}

var (
	config *configuration
	logger *logging.Logger
)

// configFunc sets the global configuration; it's overridden in tests.
var configFunc = defaultConfigFunc

type configuration struct {
	once sync.Once
	err  error
}

func (c *configuration) Error() error {
	return c.err
}

type envError struct {
	name string
}

func (e *envError) Error() string {
	return fmt.Sprintf("%s environment variable unset or missing", e.name)
}

func F(w http.ResponseWriter, r *http.Request) {
	config.once.Do(func() { configFunc() })

	defer logger.Flush()

	logger.Log(logging.Entry{
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
		Payload:  "processing HTTP request",
		Severity: logging.Info,
	})

	if config.Error() != nil {
		logger.Log(logging.Entry{
			Payload:  config.Error(),
			Severity: logging.Error,
		})
		http.Error(w, config.Error().Error(), http.StatusInternalServerError)
		return
	}

	t1()

	fmt.Fprintf(w, "Tasks completed successfully")
}

func t1() {
	logger.Log(logging.Entry{
		Payload:  "function t1 called",
		Severity: logging.Info,
	})
}

func defaultConfigFunc() {
	projectId := os.Getenv("GCP_PROJECT")
	if projectId == "" {
		config.err = &envError{"GCP_PROJECT"}
		return
	}

	functionName := os.Getenv("FUNCTION_NAME")
	if functionName == "" {
		config.err = &envError{"FUNCTION_NAME"}
		return
	}

	region := os.Getenv("FUNCTION_REGION")
	if region == "" {
		config.err = &envError{"FUNCTION_REGION"}
		return
	}

	client, err := logging.NewClient(context.Background(), projectId)
	if err != nil {
		config.err = err
		return
	}

	monitoredResource := monitoredres.MonitoredResource{
		Type: "cloud_function",
		Labels: map[string]string{
			"function_name": functionName,
			"region":        region,
		},
	}
	commonResource := logging.CommonResource(&monitoredResource)
	logger = client.Logger(functionName, commonResource)
}
