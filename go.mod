module github.com/authelia/authelia

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Gurpartap/logrus-stack v0.0.0-20170710170904-89c00d8a28f4
	github.com/Workiva/go-datastructures v1.0.53
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/deckarep/golang-set v1.7.1
	github.com/duosecurity/duo_api_golang v0.0.0-20201112143038-0e07e9f869e3
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/fasthttp/router v1.3.14
	github.com/fasthttp/session/v2 v2.3.2
	github.com/go-ldap/ldap/v3 v3.3.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/jackc/pgx/v4 v4.11.0
	github.com/knadh/koanf v1.1.0
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/mitchellh/mapstructure v1.4.1
	github.com/ory/fosite v0.40.2
	github.com/otiai10/copy v1.6.0
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pquerna/otp v1.3.0
	github.com/simia-tech/crypt v0.5.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tebeka/selenium v0.9.9
	github.com/tstranex/u2f v1.0.0
	github.com/valyala/fasthttp v1.26.0
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602 // indirect
	golang.org/x/text v0.3.6
	golang.org/x/tools v0.1.2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/mattn/go-sqlite3 v2.0.3+incompatible => github.com/mattn/go-sqlite3 v1.14.7
