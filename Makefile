docker_test: deps
	docker-compose up

deps:
	dep ensure -v

test: integration_test

unit_test:
	go test ./... -tags=unit -v

integration_test:
	go test ./... -tags=integration -v -p=1
