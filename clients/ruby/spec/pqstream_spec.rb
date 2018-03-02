# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Pqstream do
  it 'has a version number' do
    expect(Pqstream::VERSION).not_to be nil
  end
end
