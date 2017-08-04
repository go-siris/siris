#!/usr/bin/env bash
set -e
echo "" > coverage.txt
echo "#!/usr/bin/env bash" > ___source.out
for d in "$(go env)" ; do
  echo "$d" >> ___source.out
done

source ___source.out
exec make "$@" GOPATH="$GOPATH"
