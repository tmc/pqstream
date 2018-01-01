# frozen_string_literal: true

Gem::Specification.new do |s|
  s.name        = 'pqstream'
  s.version     = '0.0.1'
  # s.date        = ''
  s.summary     = ''
  s.description = ''
  s.authors     = ['']
  s.email       = ''
  s.files       = [
    'lib/pqstream.rb',
    'lib/pqstream/version.rb',
    'lib/pqstream_pb.rb',
    'lib/pqstream_services_pb.rb'
  ]
  s.homepage    = 'http://github.com/tmc/pqstream'
  s.license     = 'ISC'

  s.add_runtime_dependency 'grpc', '~> 1'
  s.add_development_dependency 'grpc-tools', '~> 1'
  s.add_development_dependency 'rubocop'
end
