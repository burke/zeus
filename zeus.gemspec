# -*- encoding: utf-8 -*-
require File.expand_path('../lib/zeus/version', __FILE__)

Gem::Specification.new do |gem|
  gem.authors       = ["Burke Libbey"]
  gem.email         = ["burke@libbey.me"]
  gem.description   = %q{Zeus preloads pretty much everything you'll ever want to use in development.}
  gem.summary       = %q{Zeus is an alpha-quality application preloader with terrible documentation.}
  gem.homepage      = "http://github.com/burke/zeus"

  gem.files         = `git ls-files`.split("\n").reject{ |f| f =~ /xcodeproj/ }
  gem.executables   = gem.files.grep(%r{^bin/}).map{ |f| File.basename(f) }
  gem.test_files    = gem.files.grep(%r{^(test|spec|features)/})
  gem.name          = "zeus"
  gem.require_paths = ["lib"]
  gem.version       = Zeus::VERSION
end
