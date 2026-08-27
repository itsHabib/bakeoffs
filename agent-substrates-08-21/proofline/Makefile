.PHONY: demo test check

demo: check
	@go run ./cmd/proofline demo
	@printf 'source lines: '
	@find cmd internal -name '*.go' ! -name '*_test.go' -print0 | xargs -0 cat | wc -l | tr -d ' '

test:
	@go test ./...

check:
	@test -z "$$(gofmt -l $$(find cmd internal -name '*.go'))"
	@(cd fixtures && shasum -a 256 -c SHA256SUMS)
	@go vet ./...
	@go test ./...
