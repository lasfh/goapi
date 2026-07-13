default: run

run:
	@docker-compose up

# Caso seja necessário instalar o swagger e/ou não encontre no path (linhas comentadas abaixo):
# go install github.com/swaggo/swag/cmd/swag@latest &&\
# export PATH=$(go env GOPATH)/bin:$PATH &&\

swagger:
	swag init -g main.go --parseDependency -o ./internal/docs

run-migrations:
	SQLSERVER_HOST=localhost \
	go run main.go migration up

setup-hooks:
	@if [ "$$(git config core.hooksPath)" != ".hooks" ]; then \
		echo "Configurando core.hooksPath para .hooks..."; \
		git config core.hooksPath .hooks; \
		echo "✓ core.hooksPath configurado para .hooks"; \
	fi

build-api-image:
	@if [ -z "$(VERSION)" ]; then \
		echo "Erro: informe a versão. Ex: make build-api-image VERSION=1.x.x"; \
		exit 1; \
	fi

	go mod tidy &&\
	docker build --build-arg VERSION=v$(VERSION) --build-arg GIT_COMMIT=$$(git rev-parse --short HEAD) -t api:$(VERSION) .


push-api-image:
	@if [ -z "$(VERSION)" ]; then \
		echo "Erro: informe a versão. Ex: make push-api-image VERSION=1.x.x"; \
		exit 1; \
	fi

	docker push api:$(VERSION)

check-api:
	go test -race -cover ./...
