all: lint

test:
	bundle exec rspec -fd
	bash -c 'bash <(curl -s https://codecov.io/bash) -F ruby'

generate:
	grpc_tools_ruby_protoc -I ../.. -I ../../vendor -I ../../../../.. --ruby_out=lib --grpc_out=lib ../../pqstream.proto
	./protoc-fixup

deps:
	bundle install

gem:
	gem build pqstream.gemspec

lint:
	rubocop
