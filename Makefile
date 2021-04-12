test:
	go test

compile_dirsearch:
	go build -o $(dest) cmd/dirsearch/main.go

run_dirsearch:
	go run cmd/dirsearch/main.go

debug_dirsearch:
	go build -o debug/dirsearch/fsac cmd/dirsearch/main.go
