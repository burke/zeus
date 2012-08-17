module Zeus
  module Plan

    class Stage < Node

      attr_reader :actions
      def initialize(name)
        super(name)
        @actions = []
      end

      def action(&b)
        @actions << b
        self
      end

      def desc(desc)
        @desc = desc
      end

      def stage(name, &b)
        @stages << Plan::Stage.new(name).tap { |s| s.instance_eval(&b) }
        self
      end

      def command(name, *aliases, &b)
        @stages << Plan::Acceptor.new(name, aliases, @desc, &b)
        @desc = nil
        self
      end

      # ^ configuration
      # V later use

      def acceptors
        stages.map(&:acceptors).flatten
      end

      def to_process_object(server)
        Zeus::Server::Stage.new(server).tap do |stage|
          stage.name = @name
          stage.stages = @stages.map { |stage| stage.to_process_object(server) }
          stage.actions = @actions
        end
      end

    end

  end
end
