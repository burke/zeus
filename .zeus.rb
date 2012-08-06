Zeus::Server.define! do

  stage :zeus do
    action { require 'zeus' ; require 'rspec' ; require 'rspec/core/runner' }

    command :spec, :s do
      RSpec::Core::Runner.autorun
    end
  end

end
