package main

//go:generate go tool xgoja build -f xgoja.yaml --work-dir . --keep-work --dry-run
//go:generate env GOWORK=off go run ./internal/postgenerate
//go:generate env GOWORK=off go mod tidy
