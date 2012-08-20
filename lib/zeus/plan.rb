require 'set'

require 'zeus/plan/node'
require 'zeus/plan/stage'
require 'zeus/plan/acceptor'

module Zeus
  module Plan
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
