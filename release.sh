#!/bin/bash

version="$1"
if [[ -z $version ]]; then
  echo "please input release version"
  exit 0
fi

git tag -a "slago-api/$version" -m "$version"
git tag -a "salgo-logrus/$version" -m "$version"
git tag -a "slago-zap/$version" -m "$version"
git tag -a "slago-zerolog/$version" -m "$version"
git tag -a "log-to-slago/$version" -m "$version"
git tag -a "logrus-to-slago/$version" -m "$version"
git tag -a "zap-to-slago/$version" -m "$version"
git tag -a "zerolog-to-slago/$version" -m "$version"

echo "release tag complete"
