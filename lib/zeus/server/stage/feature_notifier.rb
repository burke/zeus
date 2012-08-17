module Zeus
  class Server
    class Stage
      class FeatureNotifier

        def initialize(server, stage_name)
          @server = server
          @stage_name = stage_name
        end

        def notify_new_features
          new_features = newly_loaded_features()
          $previously_loaded_features ||= []
          $previously_loaded_features |= new_features
          Thread.new {
            new_features.each { |f| notify_feature(f) }
          }
        end

        def notify_feature(feature)
          @server.__CHILD__stage_has_feature(@stage_name, feature)
        end

        private

        def newly_loaded_features
          old_features = defined?($previously_loaded_features) ? $previously_loaded_features : []
          ($LOADED_FEATURES + @server.extra_features) - old_features
        end

      end
    end
  end
end




