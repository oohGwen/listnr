BINARY=build/listnr

.PHONY: build clean run

run: build
	./$(BINARY)

build:
	go build -o $(BINARY) cmd/listnr/main.go

clean:
	rm -rf build