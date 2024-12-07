package db

import (
	"github.com/google/wire"
)

var DBSet = wire.NewSet(
	ProvideDBConfig,
	ProvideDB,
)
