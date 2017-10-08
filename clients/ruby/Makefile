all: gem

generate:
	protoc --version
	protoc -I ../.. -I ../../../../.. --ruby_out=./lib --grpc_out=./lib --plugin=protoc-gen-grpc=`which grpc_ruby_plugin` ../../pqstream.proto

deps:
	bundle install

gem:
	gem build pqstream.gemspec

lint:
	rubocop
