module Zeus
  class Server
    class LoadTracking
      class << self
        attr_accessor :server

        def add_feature(file)
          return unless server
          if absolute_path?(file)
            server.add_extra_feature(file)
          elsif File.exist?("./#{file}")
            server.add_extra_feature(File.expand_path("./#{file}"))
          else
            path = find_in_load_path(file)
            server.add_extra_feature(path) if path
          end
        end

        private

        def find_in_load_path(file)
          path = $LOAD_PATH.detect { |path| File.exist?("#{path}/#{file}") }
          "#{path}/#{file}" if path
        end

        def absolute_path?(file)
          file =~ /^\// && File.exist?(file)
        end
      end
    end
  end
end

module Kernel

  def load(file, *a)
    Kernel.load(file, *a)
  end

  class << self
    alias_method :__original_load, :load
    def load(file, *a)
      Zeus::Server::LoadTracking.add_feature(file)
      __original_load(file, *a)
    end
  end

end


