module github.com/DataDog/data-streams-go

go 1.18

replace github.com/DataDog/data-streams-go/integrations/kafka => ./integrations/kafka

require (
	github.com/DataDog/datadog-go/v5 v5.3.0
	github.com/DataDog/sketches-go v1.3.0
	github.com/golang/protobuf v1.5.2
	github.com/stretchr/testify v1.8.1
	github.com/tinylib/msgp v1.1.6
)

require (
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
