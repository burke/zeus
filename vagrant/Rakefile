#!/usr/bin/env rake
require "bundler/gem_tasks"
require 'fileutils'
require 'pathname'

ROOT_PATH = Pathname.new(File.expand_path("../../", __FILE__))
RUBYGEM_PATH = Pathname.new(File.expand_path("../", __FILE__))

task build: [:manifest]
task default: :build

task :manifest do
  files = `find . -type f | sed 's|^\./||'`.lines.map(&:chomp)
  exceptions = [
    /.gitignore$/,
    /^MANIFEST$/,
    /^pkg\//,
  ]
  files.reject! { |f| exceptions.any? {|ex| f =~ ex }}
  File.open('MANIFEST', 'w') {|f| f.puts files.join("\n") }
end

