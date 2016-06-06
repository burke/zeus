module Zeus
  class LoadTracking
    class << self
      def features_loaded_by(&block)
        old_features = all_features

        # Catch exceptions so we can determine the features
        # that were being loaded at the time of the exception.
        err_features = []
        begin
          yield
        rescue SyntaxError => err
          # SyntaxErrors are a bit weird in that the file containing
          # the error is not in the backtrace, only the error message.
          match = /\A([^:]+):\d+: syntax error/.match(err.message)
          err_features << match[1] if match
        rescue Exception => err
          # Just capture this to add to the err_features list
        end

        if err && err.backtrace
          err_features += err.backtrace.map { |b| b.split(':').first }
                             .select { |f| f.start_with?('/') }
                             .take_while { |f| f != __FILE__ }
        end

        new_features = all_features + err_features - old_features

        [new_features, err]
      end

      def add_feature(file)
        add_features([file])
      end

      def add_features(files)
        files = files.map { |f| File.expand_path(f) }.select { |f| File.exist?(f) }
        Zeus.notify_features(files)
      end

      # $LOADED_FEATURES global variable is used internally by Rubygems
      def all_features
        Set.new($LOADED_FEATURES.dup)
      end
    end
  end
end


module Kernel

  def load(file, *a)
    Kernel.load(file, *a)
  end
  private :load

  class << self
    alias_method :__load_without_zeus, :load
    def load(file, *a)
      Zeus::LoadTracking.add_feature(file)
      __load_without_zeus(file, *a)
    end
  end
end

require 'yaml'
module YAML
  class << self
    alias_method :__load_file_without_zeus, :load_file
    def load_file(file, *a)
      Zeus::LoadTracking.add_feature(file)
      __load_file_without_zeus(file, *a)
    end
  end
end
