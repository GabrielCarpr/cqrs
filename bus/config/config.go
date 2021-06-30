package config

// TODO: Remove and find a better way
type Config interface {
	DBDsn() string
}
