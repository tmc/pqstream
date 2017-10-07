# frozen_string_literal: true

require 'pqstream/version'

module PQStream
  # Client allows subscribing to pqstream streams.
  class Client
    def initialize(addr)
      @addr = addr
    end
  end
end
