dev:
	go run ./cmd/app/main.go -port 80 $(ARGS)

upgrade-deps:
	go get -u ./...


heroku-log:
	heroku logs --tail

.PHONY: upgrade-deps, dev, heroku-log