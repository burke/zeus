module VagrantPlugins::Zeus
  module Config
    class Zeus < Vagrant.plugin(2, :config)
      attr_accessor :file_monitor_port

      def initialize
        @file_monitor_port = UNSET_VALUE
      end

      def finalize!
        @file_monitor_port = 7123 if @file_monitor_port == UNSET_VALUE
      end
    end
  end
end
