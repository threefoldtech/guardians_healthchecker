package spawner

// Config holds the configuration settings for the spawner tool.
type Config struct {
	Farms              []uint64     `yaml:"farms"`
	DeploymentStrategy float64      `yaml:"deployment_strategy"`
	GridEndpoints      Endpoints    `yaml:"grid_endpoints"`
	Mnemonic           string       `yaml:"mnemonic"`
	FailureStrategy    string       `yaml:"failure_strategy"`
	MaxRetries         uint64       `yaml:"max_retries"`
	SSHKey             string       `yaml:"ssh_key"`
	Influx             InfluxConfig `yaml:"influx"`
}

// Endpoints holds the URLs for grid
type Endpoints struct {
	GraphQl      string `yaml:"graphql"`
	Proxy        string `yaml:"proxy"`
	Relay        string `yaml:"relay"`
	SubstrateURL string `yaml:"substrate_url"`
}

// InfluxConfig contains the configuration settings for connecting to an InfluxDB instance.
type InfluxConfig struct {
	URL    string `yaml:"url"`
	Org    string `yaml:"org"`
	Token  string `yaml:"token"`
	Bucket string `yaml:"bucket"`
}

// vmInfo stores information about a specific VM.
type vmInfo struct {
	Farm        uint64 `json:"farm"`
	Node        uint32 `json:"node"`
	Name        string `json:"name"`
	Contract    uint64 `json:"contract"`
	ProjectName string `json:"project_name"`
}

// deploymentMetadata holds metadata for a deployment.
type deploymentMetadata struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
