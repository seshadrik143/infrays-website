# Build orchestration for the infraYS issuer + portal.
#
# Most workflows:
#   make build          — full build (SPA + Go binary). Output: bin/issuer
#   make spa            — just rebuild the React SPA
#   make go             — just rebuild the Go binary
#   make test           — go test ./...
#   make tidy           — go mod tidy + gofmt
#   make clean          — wipe build artifacts

.PHONY: build spa go test tidy clean

SPA_SRC := portal-ui
SPA_OUT := internal/portal/spa_dist

build: spa go

spa:
	cd $(SPA_SRC) && npm install --silent && npm run build
	rm -rf $(SPA_OUT)
	mkdir -p $(SPA_OUT)
	cp -r $(SPA_SRC)/dist/. $(SPA_OUT)/

go:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/issuer ./cmd/issuer
	@echo "Built bin/issuer"

test:
	go test ./...

tidy:
	go mod tidy
	gofmt -w .

clean:
	rm -rf bin/ $(SPA_OUT) $(SPA_SRC)/dist $(SPA_SRC)/node_modules
