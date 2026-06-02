package main

//go:generate go tool xgoja build -f xgoja.yaml --work-dir . --keep-work --dry-run --xgoja-replace ../../../go-go-goja
//go:generate env GOWORK=off go run ./internal/postgenerate
//go:generate env GOWORK=off go mod tidy
