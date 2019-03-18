TAG?=$(shell git rev-parse --short=8 HEAD)
export TAG

all: build-landingpage build

install: install-go-deps install-npm
install-go-deps:
	go get -u github.com/rakyll/statik
install-npm:
	npm install --no-package-lock postcss-cli purgecss cssnano autoprefixer

# -- Go related targets
test:
	go test ./...
build:
	go generate
	go build -ldflags "-X main.version=$(TAG)" -o gitreleases .

# -- Docker build
pack:
	docker build -t registry.gitlab.com/mweibel/gitreleases:$(TAG) .
push:
	docker push registry.gitlab.com/mweibel/gitreleases
apply-secret:
	kubectl apply -f k8s/secret.yml
deploy:
	./k8s/deploy.sh $(TAG)
ship: test pack push deploy

# -- Landingpage build
build-landingpage: public/style.min.css index.html public/script.js public/img/headline.png public/img/headline@2x.png public/img/logo.png public/img/favicon.ico | public public/img

public:
	mkdir -p $@
public/img:
	mkdir -p $@

public/style.min.css: style.css
	npx purgecss --css style.css --content landingpage/index.html landingpage/script.js --out .
	npx postcss style.css -o $@
	rm style.css
style.css: landingpage/gitreleases.css landingpage/tachyons.min.css
	cat landingpage/gitreleases.css landingpage/tachyons.min.css > $@

index.html:
	cat landingpage/index.html | sed 's/{{TAG}}/$(TAG)/g' > public/$@
public/%.js: landingpage/%.js
	cp $< $@
public/img/%.png: landingpage/img/%.png
	cp $< $@
public/img/%.ico: landingpage/img/%.ico
	cp $< $@

.PHONY: all install install-go-deps install-npm test build build-landingpage index.html pack push apply-secret deploy ship
