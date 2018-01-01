all: lint

test:
	CODECOV_FLAG=ruby bundle exec rspec -fd

generate:
	grpc_tools_ruby_protoc -I ../.. -I ../../vendor -I ../../../../.. --ruby_out=lib --grpc_out=lib ../../pqstream.proto
	./protoc-fixup

deps:
	bundle install

gem:
	gem build pqstream.gemspec

lint:
	rubocop
