require 'zeus/server/stage'
require 'zeus/server/acceptor'

module Zeus
  module DSL

    class Evaluator
      def stage(name, &b)
        stage = DSL::Stage.new(name)
        stage.instance_eval(&b)
      end
    end

    class Acceptor

      attr_reader :name, :aliases, :description, :action
      def initialize(name, aliases, description, &b)
        @name = name
        @description = description
        @aliases = aliases
        @action = b
      end

      # ^ configuration
      # V later use

      def acceptors
        self
      end

      def to_domain_object(server)
        Zeus::Server::Acceptor.new(server).tap do |stage|
          stage.name = @name
          stage.aliases = @aliases
          stage.action = @action
          stage.description = @description
        end
      end

    end

    class Stage

      attr_reader :pid, :stages, :actions
      def initialize(name)
        @name = name
        @stages, @actions = [], []
      end

      def action(&b)
        @actions << b
        self
      end

      def desc(desc)
        @desc = desc
      end

      def stage(name, &b)
        @stages << DSL::Stage.new(name).tap { |s| s.instance_eval(&b) }
        self
      end

      def command(name, *aliases, &b)
        @stages << DSL::Acceptor.new(name, aliases, @desc, &b)
        @desc = nil
        self
      end

      # ^ configuration
      # V later use

      def acceptors
        stages.map(&:acceptors).flatten
      end

      def to_domain_object(server)
        Zeus::Server::Stage.new(server).tap do |stage|
          stage.name = @name
          stage.stages = @stages.map { |stage| stage.to_domain_object(server) }
          stage.actions = @actions
        end
      end

    end

  end
end
