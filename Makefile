SHELL := bash
MODEL ?= llama3.2:3b

.PHONY: setup dev serve test api-port-free ollama

# One-time: frontend deps and the model.
setup:
	cd web && bun install
	ollama pull $(MODEL)

# For working on the code: live reload. Also on your network, but heavy for phones.
dev: api-port-free ollama
	@set -m; \
	(cd api && OLLAMA_MODEL=$(MODEL) go run ./cmd/server) & api=$$!; \
	(cd web && bun run dev --host) & web=$$!; \
	trap 'kill -- -$$api -$$web 2>/dev/null; (cd web && bunx astro dev stop >/dev/null 2>&1)' INT TERM EXIT; \
	wait

# For using the app, including from your phone: the production build, small and fast.
serve: api-port-free ollama
	cd web && bun run build
	@echo "On your phone (same Wi-Fi or Tailscale), open one of:"; \
	for ip in $$(hostname -I 2>/dev/null); do case $$ip in *:*|172.*) ;; *) echo "  http://$$ip:4321";; esac; done
	@set -m; \
	(cd api && OLLAMA_MODEL=$(MODEL) go run ./cmd/server) & api=$$!; \
	(cd web && HOST=0.0.0.0 PORT=4321 node server.mjs) & web=$$!; \
	trap 'kill -- -$$api -$$web 2>/dev/null' INT TERM EXIT; \
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
