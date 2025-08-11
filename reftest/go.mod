module github.com/anz-bank/decimal/reftest

go 1.24.4

require github.com/stretchr/testify v1.10.0

require (
	github.com/anz-bank/decimal v1.15.0
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/anz-bank/decimal => ..
