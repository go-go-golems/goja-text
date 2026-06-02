module github.com/go-go-golems/goja-text

go 1.26.1

toolchain go1.26.3

require (
	github.com/dop251/goja v0.0.0-20260311135729-065cd970411c
	github.com/go-go-golems/go-go-goja v0.0.0-00010101000000-000000000000
	github.com/go-go-golems/logcopter v0.1.0
	github.com/yuin/goldmark v1.8.2
)

require (
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dop251/goja_nodejs v0.0.0-20250409162600-f7acab6894b0 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/google/pprof v0.0.0-20241029153458-d1b30febd7db // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/zerolog v1.35.1 // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
)

tool github.com/go-go-golems/logcopter/cmd/logcopter-gen

replace github.com/go-go-golems/go-go-goja => ../go-go-goja
