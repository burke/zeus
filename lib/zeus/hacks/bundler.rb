begin
  require 'bundler'

  module Bundler
    class Runtime
      alias_method :setup_without_zeus, :setup
      def setup(*groups)
        ret = setup_without_zeus(*groups)
        zeus = File.expand_path("../../../", __FILE__)
        $LOAD_PATH.unshift(zeus) unless $LOAD_PATH.include?(zeus)
        ret
      end
    end
  end

rescue LoadError
end

