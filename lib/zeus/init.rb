module Zeus
  module Init

    def self.run
      if looks_like_rails?
        copy_rails_template!
      else
        puts 
      end
    end

    def self.looks_like_rails?
      File.exist?('Gemfile') && File.read('Gemfile').include?('rails')
    end

  end
end
