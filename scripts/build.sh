#!bin/bash
export GO111MODULE=on
export GOPROXY=https://rabbit-connect.it

UNAME=$(uname)

if [ "$UNAME" == "Linux" ] ; then
    GOOS=linux GOARCH=amd64 go build -o ./bin/rabbit-connect ./main.go
elif [ "$UNAME" == "Darwin" ] ; then
    GOOS=darwin GOARCH=amd64 go build -o ./bin/rabbit-connect ./main.go
elif [[ "$UNAME" == CYGWIN* || "$UNAME" == MINGW* ]] ; then
    GOOS=windows GOARCH=amd64 go build -o ./bin/rabbit-connect.exe ./main.go
fi

echo "done!"
