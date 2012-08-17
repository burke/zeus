module Zeus
  module Plan

    class Acceptor < Node

      attr_reader :name, :aliases, :description, :action
      def initialize(name, aliases, description, &b)
        super(name)
        @description = description
        @aliases = aliases
        @action = b
      end

      # ^ configuration
      # V later use

      def commands
        [name, *aliases].map(&:to_s)
      end

      def acceptors
        self
      end

      def to_process_object(server)
        Zeus::Server::Acceptor.new(server).tap do |stage|
          stage.name = @name
          stage.aliases = @aliases
          stage.action = @action
          stage.description = @description
        end
      end

    end

  end
end

