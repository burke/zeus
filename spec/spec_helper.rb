require 'zeus'

module FolderHelpers
  def write(file, content)
    ensure_folder File.dirname(file)
    File.open(file, 'w'){|f| f.write content }
  end

  def read(file)
    File.read file
  end

  def delete(file)
    `rm #{file}`
  end

  def ensure_folder(folder)
    `mkdir -p #{folder}` unless File.exist?(folder)
  end

  def root
    File.expand_path '../..', __FILE__
  end
end

RSpec.configure do |config|
  config.include FolderHelpers

  config.around do |example|
    folder = File.expand_path("../tmp", __FILE__)
    `rm -rf #{folder}`
    ensure_folder folder
    Dir.chdir folder do
      example.call
    end
    `rm -rf #{folder}`
  end
end
