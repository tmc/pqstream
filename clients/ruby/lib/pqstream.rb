# frozen_string_literal: true

require 'pqstream/version'
require 'pqstream_services_pb'

module Pqs
  module PQStream
    # Client allows subscribing to pqstream streams.
    class Client
      def initialize(addr)
        @stub = Pqs::PQStream::Stub.new(addr, :this_channel_is_insecure)
      end

      def listen(opts = { table_regexp: '.*' })
        @stub.listen(Pqs::ListenRequest.new(opts))
      end
    end
  end
end
