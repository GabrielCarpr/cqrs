module example

go 1.16

replace github.com/GabrielCarpr/cqrs => ../

require (
	github.com/GabrielCarpr/cqrs v0.0.0-00010101000000-000000000000
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/uuid v1.2.0
	github.com/jmoiron/sqlx v1.3.4
	github.com/sarulabs/di/v2 v2.4.2
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	gorm.io/driver/postgres v1.1.0
	gorm.io/gorm v1.21.10
)
