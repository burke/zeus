require "vagrant"

module VagrantPlugins
  module Zeus
    class Plugin < Vagrant.plugin("2")
      name "Zeus"

      config "zeus" do
        require_relative "config/zeus"
        VagrantPlugins::Zeus::Config::Zeus
      end

      command "zeus-file-monitor" do
        require_relative "commands/zeus-file-monitor"
        VagrantPlugins::Zeus::Commands::FileMonitor
      end
    end
  end
end
