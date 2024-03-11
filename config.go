package laydb

import "strconv"

const (
	DefaultAddr         = "127.0.0.1:8000"
	DefaultMaxKeySize   = uint32(1 * 1024)
	DefaultMaxValueSize = uint32(8 * 1024)
)

type Config struct {
	Addr string `json:"addr" toml:"addr"`
	Path string `json:"path" toml:"path"` // dir path for append-only logs
}

func (c *Config) validate() {
	if c.Addr == "" {
		c.Addr = DefaultAddr
	}
}

func DefaultConfig() *Config {
	return &Config{
		Addr: DefaultAddr,
		Path: "/tmp/laydb",
	}
}

func float64ToStr(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func strToFloat64(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}
