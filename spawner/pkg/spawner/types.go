package spawner

// Config holds the configuration settings for the spawner tool.
type Config struct {
	Farms              []uint64     `yaml:"farms"`
	DeploymentStrategy float64      `yaml:"deployment_strategy"`
	GridEndpoints      Endpoints    `yaml:"grid_endpoints"`
	Mnemonic           string       `yaml:"mnemonic"`
	FailureStrategy    string       `yaml:"failure_strategy"`
	MaxRetries         uint64       `yaml:"max_retries"`
	SSHKey             string       `yaml:"ssh_keys"`
	Influx             InfluxConfig `yaml:"influx"`
}

// Endpoints holds the URLs for grid
type Endpoints struct {
	GraphQl     string `yaml:"graphql"`
	Proxy       string `yaml:"proxy"`
	Relay       string `yaml:"relay"`
	SubsrateURL string `yaml:"subsrate_url"`
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
	Farm        uint64
	Node        uint32
	Name        string `json:"name"`
	Contract    uint64
	ProjectName string `json:"projectName"`
}

// deploymentMetadata holds metadata for a deployment.
type deploymentMetadata struct {
	Version     int    `json:"version"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	ProjectName string `json:"projectName"`
}
