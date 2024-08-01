package deployer

type Config struct {
	Farms              []uint64     `yaml:"farms"`
	DeploymentStrategy string       `yaml:"deployment_strategy"`
	GridEndpoints      Endpoints    `yaml:"grid_endpoints"`
	Mnemonic           string       `yaml:"mnemonic"`
	FailureStrategy    string       `yaml:"failure_strategy"`
	MaxRetries         uint64       `yaml:"max_retries"`
	SSHKey             string       `yaml:"ssh_keys"`
	Influx             InfluxConfig `yaml:"influx"`
}

type Endpoints struct {
	GraphQl     string `yaml:"graphql"`
	Proxy       string `yaml:"proxy"`
	Relay       string `yaml:"relay"`
	SubsrateURL string `yaml:"subsrate_url"`
}

type InfluxConfig struct {
	URL    string `yaml:"url"`
	Org    string `yaml:"org"`
	Token  string `yaml:"token"`
	Bucket string `yaml:"bucket"`
}
