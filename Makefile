main:
	docker build -t zktoro:latest -f Dockerfile .
	docker create --name zktoro zktoro:latest
	docker cp zktoro:/zktoro zktoro
	# docker rm -f build-zktoro
	chmod 755 zktoro

dev:
	docker build --tag zktoro .
	docker create --name zktoro zktoro
	go build  

tmp:
	docker create --name zktoro zktoro
	go build  