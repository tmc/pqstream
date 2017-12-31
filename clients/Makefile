all: python ruby

.PHONY: deps
deps:
	./install-protobuf
	sudo ldconfig

.PHONY: ruby
ruby:
	$(MAKE) -C ruby

.PHONY: python
python:
	$(MAKE) -C python
