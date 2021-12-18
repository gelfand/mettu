.PHONY: website
website:
	go build ./cmd/website && ./website
