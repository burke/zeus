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
        new_features.uniq!

        [new_features, err]
      end

      # Check the load path first to see if the file getting loaded is already
      # loaded. Otherwise, add the file to the $untracked_features array which
      # then gets added to $LOADED_FEATURES array.
      def add_feature(file)
        full_path = File.expand_path(file)

        if find_in_load_path(full_path) || File.exist?(full_path)
          add_extra_feature(full_path)
        end
      end

      # $LOADED_FEATURES global variable is used internally by Rubygems
      def all_features
        untracked = defined?($untracked_features) ? $untracked_features : []
        $LOADED_FEATURES + untracked
      end

      private

      def add_extra_feature(path)
        $untracked_features ||= []
        $untracked_features << path
      end

      def find_in_load_path(file_path)
        $LOAD_PATH.detect { |path| path == file_path }
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
