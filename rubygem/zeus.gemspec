# -*- encoding: utf-8 -*-

# This preamble is basically used to deal with bundler/gem_tasks, which loads the gemspec
# on rake init, even though some prerequisites are not generated until `rake build` is invoked.
version = begin
            require File.expand_path('../lib/zeus/version', __FILE__)
            Zeus::VERSION
          rescue LoadError, NameError
            "0.0.0"
          end

files = File.exist?('MANIFEST') ? File.read("MANIFEST").lines.map(&:chomp) : []

Gem::Specification.new do |gem|
  gem.authors       = ["Burke Libbey"]
  gem.email         = ["burke@libbey.me"]
  gem.description   = %q{Boot any rails app in under a second}
  gem.summary       = %q{Zeus is an intelligent preloader for ruby applications. It allows normal development tasks to be run in a fraction of a second.}
  gem.homepage      = "https://github.com/burke/zeus"

  gem.files         = files
  gem.extensions    = []
  gem.executables   = ['zeus']
  gem.test_files    = []
  gem.name          = "zeus"
  gem.require_paths = ["lib"]
  gem.version       = version
  gem.license       = "MIT"

  gem.add_runtime_dependency "method_source", ">= 0.6.7"

  gem.add_development_dependency "bundler", '~> 2.3'
end
