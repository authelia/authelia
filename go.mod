module github.com/authelia/authelia/v4

go 1.25.0

toolchain go1.26.2

require (
	authelia.com/provider/oauth2 v0.2.22
	github.com/Gurpartap/logrus-stack v0.0.0-20170710170904-89c00d8a28f4
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/authelia/jsonschema v0.1.7
	github.com/authelia/otp v1.0.4
	github.com/duosecurity/duo_api_golang v0.2.0
	github.com/fasthttp/router v1.5.4
	github.com/fasthttp/session/v2 v2.5.9
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-asn1-ber/asn1-ber v1.5.8-0.20260416181348-e7dc79048676
	github.com/go-crypt/crypt v0.4.13
	github.com/go-jose/go-jose/v4 v4.1.4
	github.com/go-ldap/ldap/v3 v3.4.13
	github.com/go-rod/rod v0.116.2
	github.com/go-sql-driver/mysql v1.9.3
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/go-webauthn/webauthn v0.16.4
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/cel-go v0.28.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-retryablehttp v0.7.8
	github.com/jackc/pgx/v5 v5.9.2
	github.com/jmoiron/sqlx v1.4.0
	github.com/knadh/koanf/parsers/yaml v1.1.0
	github.com/knadh/koanf/providers/confmap v1.0.0
	github.com/knadh/koanf/providers/env/v2 v2.0.0
	github.com/knadh/koanf/providers/posflag v1.0.1
	github.com/knadh/koanf/providers/rawbytes v1.0.0
	github.com/knadh/koanf/v2 v2.3.4
	github.com/mattn/go-sqlite3 v1.14.42
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/otiai10/copy v1.14.1
	github.com/prometheus/client_golang v1.23.2
	github.com/savsgio/gotils v0.0.0-20250924091648-bce9a52d7761
	github.com/sirupsen/logrus v1.9.4
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	github.com/trustelem/zxcvbn v1.0.1
	github.com/valyala/fasthttp v1.70.0
	github.com/weppos/publicsuffix-go v0.50.3
	github.com/wneessen/go-mail v0.7.2
	go.uber.org/mock v0.6.0
	go.yaml.in/yaml/v4 v4.0.0-rc.4
	golang.org/x/net v0.53.0
	golang.org/x/sync v0.20.0
	golang.org/x/term v0.42.0
	golang.org/x/text v0.36.0
	golang.org/x/time v0.15.0
)

require (
	cel.dev/expr v0.25.1 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/Azure/go-ntlmssp v0.1.0 // indirect
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boombuler/barcode v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto v0.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/fxamacker/cbor/v2 v2.9.1 // indirect
	github.com/go-crypt/x v0.4.14 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-webauthn/x v0.2.3 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/iancoleman/orderedmap v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/gomega v1.39.1 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/redis/go-redis/v9 v9.18.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/test-go/testify v1.1.4 // indirect
	github.com/tinylib/msgp v1.6.3 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/ysmood/fetchup v0.3.0 // indirect
	github.com/ysmood/goob v0.4.0 // indirect
	github.com/ysmood/got v0.42.3 // indirect
	github.com/ysmood/gson v0.7.3 // indirect
	github.com/ysmood/leakless v0.9.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260311181403-84a4fc48630c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260311181403-84a4fc48630c // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	filippo.io/edwards25519 v1.1.0 => filippo.io/edwards25519 v1.2.0
	golang.org/x/net => golang.org/x/net v0.53.0
)

exclude github.com/mattn/go-sqlite3 v2.0.3+incompatible
