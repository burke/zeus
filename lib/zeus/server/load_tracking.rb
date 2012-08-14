module Zeus
  class Server
    class LoadTracking
      class << self
        attr_accessor :server

        def add_feature(file)
          return unless server
          path = if File.exist?(File.expand_path(file))
            File.expand_path(file)
          else
            find_in_load_path(file)
          end
          server.add_extra_feature path if path
        end

        private

        def find_in_load_path(file)
          $LOAD_PATH.map { |path| "#{path}/#{file}" }.detect{ |file| File.exist? file }
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
    alias_method :__load_without_zeus, :load
    def load(file, *a)
      Zeus::Server::LoadTracking.add_feature(file)
      __load_without_zeus(file, *a)
    end
  end
end
