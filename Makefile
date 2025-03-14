.PHONY: build dev clean

build:
	CGO_ENABLED=1 go build -ldflags "-X main.dbDirPath=${HOME}/.local/share/tarrier"

dev:
	CGO_ENABLED=1 go build -ldflags "-X main.dbDirPath=${PWD}"

clean:
	rm tarrier *.db
