build-and-push:
	docker build --platform linux/amd64 -t sikalabs/filedrop -t ghcr.io/sikalabs/filedrop .
	docker push sikalabs/filedrop
	docker push ghcr.io/sikalabs/filedrop
