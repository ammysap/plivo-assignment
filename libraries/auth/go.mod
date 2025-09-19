module github.com/ammysap/plivo-pub-sub/libraries/auth

go 1.24.6

require (
	github.com/ammysap/plivo-pub-sub/logging v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/ilyakaznacheev/cleanenv v1.5.0
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.41.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/ammysap/plivo-pub-sub/logging => ../../logging
