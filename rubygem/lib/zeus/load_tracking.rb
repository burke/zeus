module Zeus
  class LoadTracking
    class << self
      def features_loaded_by(&block)
        old_features = all_features()
        yield
        new_features = all_features() - old_features
        return new_features
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
