.PHONY: glide deps clean

NAME := protoc-gen-gotemplate
SRCS := $(shell find . -type d -name 'vendor' -prune -o -type f -name '*.go')
export GO15VENDOREXPERIMENT=1

bin/$(NAME): $(SRCS)
	go build -o $(CURDIR)/bin/$(NAME) main.go

clean:
	rm -rf bin/*
	rm -rf vendor/*

deps: glide
	glide install

glide:
ifeq ($(shell command -v glide 2> /dev/null),)
	curl https://glide.sh/get | sh
endif
