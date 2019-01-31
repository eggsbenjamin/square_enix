docker_test:
	docker-compose up

test: integration_test

unit_test:
	go test ./... -tags=unit -v

integration_test:
	go test ./... -tags=integration -v -p=1
