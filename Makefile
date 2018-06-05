BIN:=/usr/bin
GOFILES=$(shell go list -f '{{join .GoFiles " "}}')
GOTESTFILES=$(shell go list -f '{{printf "%v %v" (join .GoFiles " ") (join .TestGoFiles " ")}}')
SYSTEMD_USER_DIR="$(HOME)/.config/systemd/user"
CONFIG_DIR="$(HOME)/.config/gsettings-upd"

build: pre-build
	go build -o .build/gsettings-upd $(GOFILES)

test:
	go test -v $(GOTESTFILES)

clean:
	rm -rf .build

pre-build:
	mkdir -p .build

install:
	cp .build/gsettings-upd $(BIN)

systemd:
	mkdir -p "$(SYSTEMD_USER_DIR)"
	mkdir -p "$(CONFIG_DIR)"
	cp gsettings-upd.service "$(SYSTEMD_USER_DIR)"
	cp config.json "$(CONFIG_DIR)"
	systemctl --user daemon-reload
	systemctl --user enable gsettings-upd

.PHONY: build test clean pre-build install systemd
