TAG?=$(shell git rev-parse --short=8 HEAD)
export TAG

test:
	go test ./...
install:
	npm install --no-package-lock postcss-cli purgecss cssnano autoprefixer
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
ship: test pack push deploy

css:
	cat landingpage/colors.css landingpage/tachyons.min.css > landingpage/style.css
purge-css: css
	npx purgecss --css landingpage/style.css --content landingpage/index.html --out .
	mv style.css landingpage/style.purged.css
min-css: purge-css
	npx postcss landingpage/style.purged.css -o landingpage/style.min.css

.PHONY: test build pack push apply-secret deploy ship
