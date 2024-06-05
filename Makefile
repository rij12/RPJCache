build:
	go build -o bin/rpjcache

run: build
	 ./bin/rpjcache

runfollower: build
	./bin/rpjcache --listenAddr :4000 --leaderAddr :3000

