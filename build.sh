#!/bin/bash
CGO_ENABLED=0 go build -tags netgo -ldflags '-w -s' -o argo-rollout-trafficrouter-example-bin .
