BIN := cache-max
CORPUS ?= $(HOME)/.claude/projects

.PHONY: demo build test vet report html clean

## demo: the hands-free judge run — assert planted fixtures, show the fixture
## report, then run the same tool on 30 days of real transcripts.
demo: build
	@echo "=================================================================="
	@echo " 1/3  POLICY TESTS — exact hit ratios, bust counts, dollar deltas"
	@echo "=================================================================="
	go test ./...
	@echo ""
	@echo "=================================================================="
	@echo " 2/3  FIXTURES — one steady session, one with two planted busts"
	@echo "=================================================================="
	./$(BIN) -dir fixtures -sessions 5
	@echo ""
	@echo "=================================================================="
	@echo " 3/3  REAL CORPUS — 30 days of this machine's transcripts"
	@echo "=================================================================="
	./$(BIN) -dir "$(CORPUS)" -sessions 8

build:
	go build -o $(BIN) ./cmd/cache-max

test:
	go test ./...

vet:
	gofmt -l . && go vet ./...

## report: full report over the real corpus.
report: build
	./$(BIN) -dir "$(CORPUS)" -sessions 15

## html: also emit the static single-file HTML report.
html: build
	./$(BIN) -dir "$(CORPUS)" -html report.html

clean:
	rm -f $(BIN) report.html
