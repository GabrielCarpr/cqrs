module example

go 1.16

replace github.com/GabrielCarpr/cqrs => ../

replace github.com/GabrielCarpr/cqrs/gen => ../gen

require (
	github.com/GabrielCarpr/cqrs v0.2.0
	github.com/gin-gonic/gin v1.7.2
	github.com/golang-migrate/migrate/v4 v4.2.0
	github.com/google/uuid v1.2.0
	github.com/jackc/pgconn v1.8.1
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.2
	github.com/mitchellh/mapstructure v1.4.1
	github.com/sarulabs/di/v2 v2.4.2
	github.com/stretchr/testify v1.7.0
	github.com/thoas/go-funk v0.8.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/postgres v1.1.0
	gorm.io/gorm v1.21.10
)
