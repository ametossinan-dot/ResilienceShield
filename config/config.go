package config

type Config struct {
	Port          string
	RemoteDBUrl   string
	CheckInterval int
	LocalDBPath   string
	RemoteHost    string
}

func Load() *Config {
	return &Config{
		Port:          "8080",
		RemoteDBUrl:   "http://localhost:3000/api",
		CheckInterval: 5,
		LocalDBPath:   "./data/resilienceshield.db",
		RemoteHost:    "localhost:3000",
	}
}
