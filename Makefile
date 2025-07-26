.PHONY: dvs

build:
	./build

pack:
	go run scripts/pack.go

dvs:
	go run scripts/pack.go && ./build