all: gem

deps:
	gem install rubocop

gem:
	gem build pqstream.gemspec

lint:
	rubocop
