module github.com/xxnuo/gobergamot

go 1.24.5

require (
	github.com/jerbob92/wazero-emscripten-embind v1.5.2
	github.com/tetratelabs/wazero v1.9.0
	sigs.k8s.io/yaml v1.6.0
)

require (
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/text v0.15.0 // indirect
)

retract (
	v0.1.2 // contains only retractions
	v0.1.1 // contains invalid module name
)
