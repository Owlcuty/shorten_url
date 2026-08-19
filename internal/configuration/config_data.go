package configuration

type HasherConfig struct {
	Generator string `json:"generator"`
	Encoder   string `json:"encoder"`
}

type DatabaseConfig struct {
	Type         string `json:"type"`
	User         string `json:"username"`
	Password     string `json:"password"`
	Address      string `json:"address"`
	Port         string `json:"port"`
	DatabaseName string `json:"database_name"`
}

type CacheConfig struct {
	Type     string `json:"type"`
	User     string `json:"username"`
	Password string `json:"password"`
	Address  string `json:"address"`
	Port     string `json:"port"`
}

type RepositoryConfig struct {
	DB    DatabaseConfig `json:"database"`
	Cache CacheConfig    `json:"cache"`
}

type APIConfig struct {
	Shorten string `json:"shorten_post"`
	Resolve string `json:"resolve_get"`
	Stat    string `json:"stat_get"`
}

type HttpConfig struct {
	Address string    `json:"address"`
	Port    string    `json:"port"`
	API     APIConfig `json:"api"`
}

type Config struct {
	Repo   RepositoryConfig `json:"repository"`
	Http   HttpConfig       `json:"http"`
	Hasher HasherConfig     `json:"hasher"`
}
