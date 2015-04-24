.PHONY: build

all: build

bootstrap:
	@gox -build-toolchain

setup:
	@mkdir build bin || true
	go get -u github.com/mitchellh/gox

self-build:
	@bunch go build -o bin/bunch . 

build:
	@go get github.com/dkulchenko/bunch
	@go build -o bin/bunch .

run:
	@bin/bunch

clean:
	@rm bin/bunch >/dev/null 2>&1 || true

fullclean: clean
	@rm -fr build/*

create-zip:
	@mkdir -p build/bunch
	@mv bunch_$(build_os)$(dest_ext) build/bunch/bunch$(dest_ext)
	@cp README.md build/bunch/README
	@cd build && zip -r bunch_$(build_os).zip bunch
	@rm -r build/bunch

build-linux: clean
	@gox -osarch="linux/386"
	@gox -osarch="linux/amd64"
	@$(MAKE) create-zip build_os=linux_386
	@$(MAKE) create-zip build_os=linux_amd64

build-osx: clean
	@gox -os="darwin"
	@$(MAKE) create-zip build_os=darwin_386
	@$(MAKE) create-zip build_os=darwin_amd64

build-windows: clean
	@gox -os="windows"
	@$(MAKE) create-zip build_os=windows_386 dest_ext=.exe
	@$(MAKE) create-zip build_os=windows_amd64 dest_ext=.exe

build-all: fullclean build-linux build-windows build-osx clean
