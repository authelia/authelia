module github.com/authelia/authelia/v4

go 1.21

require (
	authelia.com/provider/oauth2 v0.1.3
	github.com/Gurpartap/logrus-stack v0.0.0-20170710170904-89c00d8a28f4
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/authelia/jsonschema v0.1.7
	github.com/authelia/otp v1.0.0
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/duosecurity/duo_api_golang v0.0.0-20240205144049-bb361ad4ae1c
	github.com/fasthttp/router v1.5.0
	github.com/fasthttp/session/v2 v2.5.4
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-asn1-ber/asn1-ber v1.5.5
	github.com/go-crypt/crypt v0.2.19
	github.com/go-jose/go-jose/v4 v4.0.1
	github.com/go-ldap/ldap/v3 v3.4.6
	github.com/go-rod/rod v0.114.8
	github.com/go-sql-driver/mysql v1.8.1
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1
	github.com/go-webauthn/webauthn v0.10.2
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-retryablehttp v0.7.5
	github.com/jackc/pgx/v5 v5.5.5
	github.com/jmoiron/sqlx v1.3.5
	github.com/knadh/koanf/parsers/yaml v0.1.0
	github.com/knadh/koanf/providers/confmap v0.1.0
	github.com/knadh/koanf/providers/env v0.1.0
	github.com/knadh/koanf/providers/posflag v0.1.0
	github.com/knadh/koanf/providers/rawbytes v0.1.0
	github.com/knadh/koanf/v2 v2.1.0
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/ory/herodot v0.10.3-0.20230807143059-27cd6936499b
	github.com/ory/x v0.0.625
	github.com/otiai10/copy v1.14.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.19.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	github.com/trustelem/zxcvbn v1.0.1
	github.com/valyala/fasthttp v1.52.0
	github.com/wneessen/go-mail v0.4.1
	go.uber.org/mock v0.4.0
	golang.org/x/net v0.22.0
	golang.org/x/sync v0.6.0
	golang.org/x/term v0.18.0
	golang.org/x/text v0.14.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/fxamacker/cbor/v2 v2.6.0 // indirect
	github.com/go-crypt/x v0.2.13 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-webauthn/x v0.1.9 // indirect
	github.com/golang/glog v1.2.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-tpm v0.9.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/iancoleman/orderedmap v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jandelgado/gcov2lcov v1.0.5 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/onsi/gomega v1.27.10 // indirect
	github.com/ory/go-acc v0.2.9-0.20230103102148-6b1c9a70dbbe // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/redis/go-redis/v9 v9.5.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/savsgio/gotils v0.0.0-20240303185622-093b76447511 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.16.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/test-go/testify v1.1.4 // indirect
	github.com/tinylib/msgp v1.1.9 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/ysmood/fetchup v0.2.3 // indirect
	github.com/ysmood/goob v0.4.0 // indirect
	github.com/ysmood/got v0.34.1 // indirect
	github.com/ysmood/gson v0.7.3 // indirect
	github.com/ysmood/leakless v0.8.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

exclude github.com/mattn/go-sqlite3 v2.0.3+incompatible
