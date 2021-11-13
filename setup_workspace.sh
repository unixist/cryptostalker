path="$HOME/workspace.$RANDOM"
export GOPATH=$path
export GOBIN=$path/bin
mkdir -p $path/src
cd $path/src
go get github.com/unixist/cryptostalker
go install github.com/unixist/cryptostalker
echo -e 'Now you can run:\n  $GOBIN/cryptostalker /tmp'
