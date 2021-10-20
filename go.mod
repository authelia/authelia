module github.com/authelia/authelia/v4

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Gurpartap/logrus-stack v0.0.0-20170710170904-89c00d8a28f4
	github.com/Workiva/go-datastructures v1.0.53
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/deckarep/golang-set v1.7.1
	github.com/duosecurity/duo_api_golang v0.0.0-20211012165811-2e1863933615
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/fasthttp/router v1.4.4
	github.com/fasthttp/session/v2 v2.4.4
	github.com/go-ldap/ldap/v3 v3.4.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/jackc/pgx/v4 v4.13.0
	github.com/knadh/koanf v1.3.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/mitchellh/mapstructure v1.4.2
	github.com/ory/fosite v0.40.2
	github.com/ory/herodot v0.9.12
	github.com/otiai10/copy v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.3.0
	github.com/simia-tech/crypt v0.5.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/tebeka/selenium v0.9.9
	github.com/tstranex/u2f v1.0.0
	github.com/valyala/fasthttp v1.31.0
	golang.org/x/sys v0.0.0-20210902050250-f475640dd07b // indirect
	golang.org/x/text v0.3.7
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/mattn/go-sqlite3 v2.0.3+incompatible => github.com/mattn/go-sqlite3 v1.14.8
