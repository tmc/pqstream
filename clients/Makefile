all: python ruby

.PHONY: deps
deps:
	./install-protobuf
	sudo ldconfig

.PHONY: python
python:
	$(MAKE) -C python

.PHONY: ruby
ruby:
	$(MAKE) -C ruby
