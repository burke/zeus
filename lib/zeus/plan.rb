require 'set'

module Zeus
  module Plan
    autoload :Node,     'zeus/plan/node'
    autoload :Stage,    'zeus/plan/stage'
    autoload :Acceptor, 'zeus/plan/acceptor'

    class Evaluator
      def stage(name, &b)
        stage = Plan::Stage.new(name)
        stage.root = true
        stage.instance_eval(&b)
        stage
      end
    end

  end
end
