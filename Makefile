test:
	go test

compile_dirf:
	go build -o $(dest) cmd/dirf/main.go

run_dirf:
	go run cmd/dirf/main.go

debug_dirf:
	go build -o debug/dirf/fsac cmd/dirf/main.go
