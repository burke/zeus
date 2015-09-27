# -*- encoding: utf-8 -*-

# This preamble is basically used to deal with bundler/gem_tasks, which loads the gemspec
# on rake init, even though some prerequisites are not generated until `rake build` is invoked.
version = begin
            require File.expand_path('../lib/vagrant-zeus/version', __FILE__)
            VagrantPlugins::Zeus::VERSION
          rescue LoadError
            "0.0.0"
          end

files = File.exist?('MANIFEST') ? File.read("MANIFEST").lines.map(&:chomp) : []

Gem::Specification.new do |gem|
  gem.authors       = ["Burke Libbey"]
  gem.email         = ["burke@libbey.me"]
  gem.description   = %q{Vagrant plugin to pass along filesystem events on directories shared with the VM}
  gem.summary       = %q{This plugin watches for filesystem events on the local filesystem, and sends them over a network socket to Zeus listening from inside the VM.}
  gem.homepage      = "http://zeus.is"

  gem.files         = files
  gem.extensions    = [
    "ext/inotify-wrapper/extconf.rb",
  ]
  gem.executables   = []
  gem.test_files    = []
  gem.name          = "vagrant-zeus"
  gem.require_paths = ["lib"]
  gem.version       = version
  gem.license       = "MIT"
end
