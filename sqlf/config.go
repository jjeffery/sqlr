package sqlf

import (
	"github.com/jjeffery/sqlf/private/colname"
	"github.com/jjeffery/sqlf/private/dialect"
)

// DefaultConfig contains the default configuration, which can be
// overridden by the calling program.
var DefaultConfig *Config = NewConfig()

// Config contains configuration information that affects the generated SQL.
type Config struct {
	// Dialect used for constructing SQL queries. If nil, the dialect
	// is inferred from the list of SQL drivers loaded in the program.
	Dialect Dialect

	// Convention contains methods for inferring the name
	// of database columns from the associated Go struct field names.
	Convention Convention
}

// NewConfig returns a new Config with default values.
func NewConfig() *Config {
	return &Config{}
}

func (c *Config) clone() *Config {
	if c == nil {
		return NewConfig()
	}
	config := *c
	return &config
}

// WithDialect returns a copy of c with the dialect set.
func (c *Config) WithDialect(dialect Dialect) *Config {
	c2 := c.clone()
	c2.Dialect = dialect
	return c2
}

// WithConvention returns a copy of c with the convention set.
func (c *Config) WithConvention(convention Convention) *Config {
	c2 := c.clone()
	c2.Convention = convention
	return c2
}

// merge returns a copy of c with config from the others merged in.
func (c *Config) merge(others ...*Config) *Config {
	c2 := c.clone()
	for _, other := range others {
		if other != nil {
			if other.Dialect != nil {
				c2.Dialect = other.Dialect
			}
			if other.Convention != nil {
				c2.Convention = other.Convention
			}
		}
	}
	return c2
}

func (cfg *Config) dialect() dialect.Dialect {
	if cfg.Dialect != nil {
		return cfg.Dialect
	}
	if DefaultConfig.Dialect != nil {
		return DefaultConfig.Dialect
	}
	return dialect.New("")
}

func (cfg *Config) convention() Convention {
	if cfg.Convention != nil {
		return cfg.Convention
	}
	if DefaultConfig.Convention != nil {
		return DefaultConfig.Convention
	}
	return colname.Snake
}
