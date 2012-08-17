module Zeus
  class Server
    class Stage

      module ErrorState
        def handle_load_error(e)
          errored_file = full_path_of_file_from_error(e)

          # register all the decendent acceptors as stubs with errors
          register_acceptors_as_errors(e)

          feature_notifier.notify_feature(errored_file)
          feature_notifier.notify_new_features

          # we do not need to do anything. We wait, until a dependency changes.
          # At that point, we get killed and restarted.
          sleep
        end

        private

        def full_path_of_file_from_error(e)
          errored_file = e.backtrace[0].scan(/(.+?):\d+:in/)[0][0]

          # handle relative paths
          unless errored_file =~ /^\//
            errored_file = File.expand_path(errored_file, Dir.pwd)
          end
        end

        def register_acceptors_as_errors(e)
          descendent_acceptors.each do |acc|
            acc = acc.extend(Acceptor::ErrorState)
            acc.error = e
            acc.run
          end
        end
      end

    end
  end
end
