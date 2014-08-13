require './lib/zeus/version'

Gem::Specification.new do |gem|
  gem.authors       = ["Burke Libbey"]
  gem.email         = ["burke@libbey.me"]
  gem.description   = "Boot any rails app in under a second"
  gem.summary       = "Zeus is an intelligent preloader for ruby applications. It allows normal development tasks to be run in a fraction of a second."
  gem.homepage      = "http://zeus.is"

  gem.files         = Dir["{lib,ext,bin}/**/*"]
  gem.extensions    = ["ext/inotify-wrapper/extconf.rb"]
  gem.executables   = ['zeus']
  gem.name          = "zeus"
  gem.version       = Zeus::VERSION
  gem.license       = "MIT"

  gem.add_development_dependency "rspec", '~> 2.12.0'
  gem.add_development_dependency "rake"
  gem.add_development_dependency "ronn", '>= 0.7.0'

  gem.add_runtime_dependency "method_source", ">= 0.6.7"
end
