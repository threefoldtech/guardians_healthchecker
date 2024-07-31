package deployer

type Config struct {
	Farms              []uint64  `yaml:"farms"`
	DeploymentStrategy string    `yaml:"deployment_strategy"`
	GridEndpoints      Endpoints `yaml:"grid_endpoints"`
	Mnemonic           string    `yaml:"mnemonic"`
	FailureStrategy    string    `yaml:"failure_strategy"`
	MaxRetries         uint64    `yaml:"max_retries" json:"max_retries"`
}

type Endpoints struct {
	GraphQl     string `yaml:"graphql"`
	Proxy       string `yaml:"proxy"`
	Relay       string `yaml:"relay"`
	SubsrateURL string `yaml:"subsrate_url"`
}
