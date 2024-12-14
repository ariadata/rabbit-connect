.PHONY: run build git-update clean

run:
	go run ./main.go

build:
	CGO_ENABLED=0 go build -o rabbit-connect main.go

git-update:
	git add .
	git commit -am "update"
	git push

clean:
	rm -f rabbit-connect
