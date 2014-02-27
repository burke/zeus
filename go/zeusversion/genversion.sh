#!/bin/bash

VERSION=$(cat ../../VERSION)

cat <<EOS > zeusversion.go
package zeusversion

const VERSION string = "$VERSION"
EOS
