.PHONY: demo

demo:
	@(cd fixtures && shasum -a 256 -c SHA256SUMS)
	@test -z "$$(gofmt -l $$(find . -name '*.go' -type f))"
	@go vet ./...
	@go test ./...
	@go run ./cmd/branchroom
	@printf '\nweighted source lines: '
	@find . -name '*.go' -type f ! -name '*_test.go' -print0 | xargs -0 wc -l | tail -1 | awk '{print $$1}'
