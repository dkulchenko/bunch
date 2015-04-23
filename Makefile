.PHONY: build

all: clean build

bootstrap:
	@gox -build-toolchain

setup:
	@mkdir build || true
	go get -u github.com/mitchellh/gox

build:
	@go build .

run:
	@./bunch

clean:
	@rm bunch >/dev/null 2>&1 || true

fullclean: clean
	@rm -fr build/*

create-zip:
	@mkdir -p build/bunch
	@mv bunch_$(build_os)$(dest_ext) build/bunch/bunch$(dest_ext)
	@cp README.md build/bunch/README
	@cp conf/example.yml build/bunch/example-config.yml
	@cd build && zip -r bunch_$(build_os).zip bunch
	@rm -r build/bunch

build-linux: clean
	@go-bindata -prefix sqlite-bin/linux/ sqlite-bin/linux/
	@gox -osarch="linux/386"
	@gox -osarch="linux/amd64"
	@$(MAKE) create-zip build_os=linux_386
	@$(MAKE) create-zip build_os=linux_amd64

build-osx: clean
	@go-bindata -prefix sqlite-bin/osx/ sqlite-bin/osx/
	@gox -os="darwin"
	@$(MAKE) create-zip build_os=darwin_386
	@$(MAKE) create-zip build_os=darwin_amd64

build-windows: clean
	@go-bindata -prefix sqlite-bin/windows/ sqlite-bin/windows/
	@gox -os="windows"
	@$(MAKE) create-zip build_os=windows_386 dest_ext=.exe
	@$(MAKE) create-zip build_os=windows_amd64 dest_ext=.exe

build-all: fullclean build-linux build-windows build-osx clean
