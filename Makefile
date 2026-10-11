SHELL := bash
MODEL ?= llama3.2:3b

.PHONY: setup dev test api-port-free ollama

# One-time: frontend deps and the model.
setup:
	cd web && bun install
	ollama pull $(MODEL)

# Start Ollama (if needed), the API, and the frontend. Ctrl-C stops them all.
dev: api-port-free ollama
	@set -m; \
	(cd api && OLLAMA_MODEL=$(MODEL) go run ./cmd/server) & api=$$!; \
	(cd web && bun run dev) & web=$$!; \
	trap 'kill -- -$$api -$$web 2>/dev/null; (cd web && bunx astro dev stop >/dev/null 2>&1)' INT TERM EXIT; \
	wait

test:
	cd api && go test ./...
	cd web && bunx astro check

api-port-free:
	@if (echo >/dev/tcp/127.0.0.1/8080) 2>/dev/null; then \
		echo "Port 8080 is already in use (another API still running?). Stop it first: kill \$$(lsof -ti:8080)"; \
		exit 1; \
	fi

ollama:
	@curl -sf localhost:11434/api/version >/dev/null || (ollama serve >/tmp/sidequestar-ollama.log 2>&1 &)
