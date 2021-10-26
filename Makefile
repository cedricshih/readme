all: readme

readme:
	go build -o $@ ./cmd/readme

.PHONY: readme

