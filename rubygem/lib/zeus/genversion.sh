#!/bin/sh

cat <<EOS > version.rb
module Zeus
	VERSION = "$(cat ../../../VERSION)"
end
EOS
