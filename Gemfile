source 'https://rubygems.org'

group :test do
  gem 'rspec'
end

# Only this group is skipped on Travis (--without development)
group :development do
  gem 'rake' # called from ./man dir
  gem 'ronn'
  gem 'mustache', '=0.7.0', '<1.0' if RUBY_VERSION < '2.0'
end
