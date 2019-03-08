TAG?=$(shell git rev-parse --short=8 HEAD)
export TAG

test:
	go test ./...
build:
	go build -ldflags "-X main.version=$(TAG)" -o gitreleases .
pack:
	docker build -t registry.gitlab.com/mweibel/gitreleases:$(TAG) .
push:
	docker push registry.gitlab.com/mweibel/gitreleases
apply-secret:
	kubectl apply -f k8s/secret.yml
deploy:
	cat k8s/deployment.yml | sed 's/{{TAG}}/$(TAG)/g'| kubectl apply -f -
ship: test pack upload deploy

.PHONY: test build pack push apply-secret deploy ship
