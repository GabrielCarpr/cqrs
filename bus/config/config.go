package config

type Config interface {
	DBDsn() string
}
