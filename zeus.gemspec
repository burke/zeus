# -*- encoding: utf-8 -*-
require File.expand_path('../lib/zeus/version', __FILE__)

Gem::Specification.new do |gem|
  gem.authors       = ["Burke Libbey"]
  gem.email         = ["burke@libbey.me"]
  gem.description   = %q{Boot any rails app in under a second}
  gem.summary       = %q{Zeus is an intelligent preloader for ruby applications. It allows normal development tasks to be run in a fraction of a second.}
  gem.homepage      = "https://github.com/burke/zeus"

  gem.files         = `git ls-files`.split("\n").reject{ |f| f =~ /xcodeproj/ }
  gem.executables   = gem.files.grep(%r{^bin/}).map{ |f| File.basename(f) }
  gem.test_files    = gem.files.grep(%r{^(test|spec|features)/})
  gem.name          = "zeus"
  gem.require_paths = ["lib"]
  gem.version       = Zeus::VERSION
  gem.license       = "MIT"
end
