# frozen_string_literal: true

require 'pqstream'

c = Pqs::PQStream::Client.new('localhost:7000')
c.listen.each { |x| puts x.to_json }
