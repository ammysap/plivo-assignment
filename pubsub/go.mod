module github.com/ammysap/plivo-pub-sub/pubsub

go 1.24.6

require (
	github.com/ammysap/plivo-pub-sub/logging v0.0.0
	github.com/google/uuid v1.6.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)

replace github.com/ammysap/plivo-pub-sub/logging => ../logging
