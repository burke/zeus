module Zeus
  class Server
    class LoadTracking

      def self.inject!(server)
        mod = module_for_server(server)
        Object.class_eval { include mod }
      end

      def self.module_for_server(server)
        Module.new.tap do |load_tracking|
          load_tracking.send(:define_method, :load) { |file, *a|
            if ret = super(file, *a)
              LoadTracking.add_feature(server, file)
            end
            ret
          }
        end
      end

      def self.add_feature(server, file)
        if absolute_path?(file)
          server.add_extra_feature(file)
        elsif File.exist?("./#{file}")
          server.add_extra_feature(File.expand_path("./#{file}"))
        else
          path = find_in_load_path(file)
          server.add_extra_feature(path) if path
        end
      end

      def self.find_in_load_path(file)
        path = $LOAD_PATH.detect { |path| File.exist?("#{path}/#{file}") }
        "#{path}/#{file}" if path
      end

      def self.absolute_path?(file)
        file =~ /^\// && File.exist?(file)
      end

    end
  end
end
