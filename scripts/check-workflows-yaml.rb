#!/usr/bin/env ruby
# Valida sintaxis YAML de workflows de GitHub Actions.

require "yaml"

Dir[File.expand_path("../.github/workflows/*.yml", __dir__)].sort.each do |file|
  YAML.load_file(file)
end

puts "yaml ok"
