# Without a Metro API token, local runs take their live data from the deployed server.
# Until api.caislx.tech is live, run with METRO_BACKEND_URL=https://api-olhometrolx.duckdns.org
METRO_BACKEND_URL ?= https://api.caislx.tech
export METRO_BACKEND_URL

.PHONY: run dev-api dev-web frontend server mac icons clean

## run: build the frontend and serve the whole app on http://localhost:8080
run: frontend
	go run .

## dev-api: the Go API alone on :8080 (pair it with dev-web)
dev-api:
	go run .

## dev-web: Vite with hot reload on http://localhost:5173, /api proxied to :8080
dev-web: frontend/node_modules
	cd frontend && npm run dev

## frontend: build the frontend into web/dist, where the Go binary embeds it
frontend: frontend/node_modules
	cd frontend && npm run build

frontend/node_modules: frontend/package-lock.json
	cd frontend && npm ci
	touch $@

## server: the Linux binary the deploy ships
server: frontend
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/metro-lis-go .

## mac: build/Cais.app for macOS
mac: frontend
	./scripts/build-macos.sh

## icons: redraw the PWA and iPhone icons in frontend/public
icons:
	go run ./tools/icons -pwa frontend/public

clean:
	rm -rf build
	find web/dist -mindepth 1 ! -name .gitkeep -delete
