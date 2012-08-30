# -*- encoding: utf-8 -*-
begin
  require File.expand_path('../lib/zeus/version', __FILE__)
rescue LoadError
  # this happens if the `version` rake task has not been run.
end

Gem::Specification.new do |gem|
  gem.authors       = ["Burke Libbey"]
  gem.email         = ["burke@libbey.me"]
  gem.description   = %q{Boot any rails app in under a second}
  gem.summary       = %q{Zeus is an intelligent preloader for ruby applications. It allows normal development tasks to be run in a fraction of a second.}
  gem.homepage      = "http://zeus.is"

  gem.files         = File.read('MANIFEST').lines.map(&:chomp)
  gem.executables   = ['zeus']
  gem.test_files    = []
  gem.name          = "zeus"
  gem.require_paths = ["lib"]
  gem.version       = defined?(Zeus::VERSION) ? Zeus::VERSION : "0.0.0"
  gem.license       = "MIT"
end
