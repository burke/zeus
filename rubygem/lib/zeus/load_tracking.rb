module Zeus
  class LoadTracking
    class << self
      def add_feature(file)
        full_path = File.expand_path(file)
        return unless File.exist?(full_path) && @feature_pipe
        notify_features([full_path])
      end

      # Internal: This should only be called by Zeus code
      def track_features_loaded_by
        old_features = $LOADED_FEATURES.dup

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
          raise
        rescue ScriptError => err
          raise
        rescue => err
          raise
        ensure
          if err && err.backtrace
            err_features += err.backtrace.map { |b| b.split(':').first }
                           .select { |f| f.start_with?('/') }
                           .take_while { |f| f != __FILE__ }
          end

          notify_features(Set.new($LOADED_FEATURES) + err_features - old_features)
        end
      end

      # Internal: This should only be called by Zeus code
      def set_feature_pipe(feature_pipe)
        @feature_mutex = Mutex.new
        @feature_pipe = feature_pipe
      end

      # Internal: This should only be called by Zeus code
      def clear_feature_pipe
        @feature_pipe.close
        @feature_pipe = nil
        @feature_mutex = nil
      end

      private

      def notify_features(features)
        unless @feature_pipe
          raise "Attempted to report features to Zeus when not running as part of a Zeus process"
        end

        @feature_mutex.synchronize do
          features.each do |t|
            @feature_pipe.puts(t)
          end
        end
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
