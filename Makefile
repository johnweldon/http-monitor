.PHONY: run
run: gen
	go run .

.PHONY: gen
gen:
	go generate

