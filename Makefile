build:
	go build -o bin/rpjcache

run: build
	 ./bin/rpjcache

runfollower: build
	./bin/rpjcache --listenaddr :4000 --leaderaddr :3000

test:
	go test -v ./...

