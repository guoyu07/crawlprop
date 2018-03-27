package config

var Version string
var Configuration interface{}

type Config struct {
	Logging LoggingConfig `toml:"logging" json:"logging"`
	Api     ApiConfig     `toml:"api" json:"api"`
}

type ApiConfig struct {
	Enabled bool   `toml:"enabled" json:"enabled"`
	Bind    string `toml:"bind" json:"bind"`
}

type LoggingConfig struct {
	Level  string `toml:"level" json:"level"`
	Output string `toml:"output" json:"output"`
}
