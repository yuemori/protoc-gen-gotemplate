PROTO := $(shell find . -name '*.proto')
.PHONY: $(PROTO) all gen

all: gen

$(PROTO):
	protoc --gotemplate_out=templates/mytemplate.md.tmpl:output \
		--plugin=protoc-gen-gotemplate=$(CURDIR)/../bin/protoc-gen-gotemplate \
		--proto_path=proto \
		$@

gen:
	protoc --gotemplate_out=templates/mytemplate.md.tmpl:output \
		--plugin=protoc-gen-gotemplate=$(CURDIR)/../bin/protoc-gen-gotemplate \
		--proto_path=proto \
		$(PROTO)
