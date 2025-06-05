package mongodb

import (
	"fmt"
	"time"
)

type Config struct {
	Host        string
	Port        int
	Database    string
	Username    string
	Password    string
	AuthSource  string
	MaxPoolSize uint64
	MinPoolSize uint64
	MaxIdleTime time.Duration
	Timeout     time.Duration
	SSL         bool
	ReplicaSet  string
}

func NewConfig() *Config {
	return &Config{
		Host:        "localhost",
		Port:        27017,
		Database:    "test",
		AuthSource:  "admin",
		MaxPoolSize: 100,
		MinPoolSize: 5,
		MaxIdleTime: 30 * time.Second,
		Timeout:     10 * time.Second,
		SSL:         false,
	}
}

func (c *Config) ConnectionString() string {
	uri := "mongodb://"

	if c.Username != "" && c.Password != "" {
		uri += fmt.Sprintf("%s:%s@", c.Username, c.Password)
	}

	uri += fmt.Sprintf("%s:%d/%s", c.Host, c.Port, c.Database)

	params := make([]string, 0)

	if c.AuthSource != "" && c.Username != "" {
		params = append(params, fmt.Sprintf("authSource=%s", c.AuthSource))
	}

	if c.MaxPoolSize > 0 {
		params = append(params, fmt.Sprintf("maxPoolSize=%d", c.MaxPoolSize))
	}

	if c.MinPoolSize > 0 {
		params = append(params, fmt.Sprintf("minPoolSize=%d", c.MinPoolSize))
	}

	if c.SSL {
		params = append(params, "ssl=true")
	}

	if c.ReplicaSet != "" {
		params = append(params, fmt.Sprintf("replicaSet=%s", c.ReplicaSet))
	}

	if len(params) > 0 {
		uri += "?"
		for i, param := range params {
			if i > 0 {
				uri += "&"
			}
			uri += param
		}
	}

	return uri
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.Database == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	return nil
}
