# Spawner


A command line tool used for spawning and destroying benchmark VMs.

## Installation
1. **Clone the repository:**
``` bash
git clone github.com/threefoldtech/guardians_healthchecker
```
2. **Navigate to spawner directory:**
``` bash
cd spawner
```

## Build
Inside the spawner directory run the following command:
``` bash
make build
```
Move the binary to any of `$PATH` directories, for example:
``` bash
mv spawner /usr/local/bin
```

## Supported Configurations
Create a new configuration file, for example `config.yaml`

| Field                  | Description                                          | Supported Values                                     | Required |
| ---------------------- | ---------------------------------------------------- | ---------------------------------------------------- | -------- |
| `farms`                | List of farm IDs where VMs should be deployed        | List of integers (e.g., `1`, `2`, etc.)              | Yes      |
| `deployment_strategy`  | Strategy for deploying VMs across nodes              | `1`, `0.7`, `0.5`, etc.                              | Yes      |
| `grid_endpoints`       | URLs for grid services                               |                                                      |      |
| `grid_endpoints.graphql` | GraphQL endpoint URL                               | URL (e.g., `"https://graphql.dev.grid.tf/graphql"`)  | Yes      |
| `grid_endpoints.proxy`   | Proxy endpoint URL                                 | URL (e.g., `"https://gridproxy.dev.grid.tf/"`)       | Yes      |
| `grid_endpoints.relay`   | Relay endpoint URL                                 | URL (e.g., `"wss://relay.dev.grid.tf"`)              | Yes      |
| `grid_endpoints.substrate_url` | Substrate URL                                | URL (e.g., `"wss://tfchain.dev.grid.tf/ws"`)         | Yes      |
| `failure_strategy`     | Strategy for handling deployment failures            | `"retry"`, `"stop"`, `"destroy-all"`, `"destroy-failing"` | Yes       |
| `mnemonic`             | Mnemonic for authentication                          | String                                               | Yes      |
| `ssh_key`              | SSH key for accessing VMs                            | String                                               | No       |
| `influx`               | InfluxDB configuration                               |                                                      |          |
| `influx.url`           | InfluxDB URL                                         | URL (e.g., `"http://influxdb.example.com"`)          | Yes      |
| `influx.org`           | InfluxDB organization name                           | String                                               | Yes      |
| `influx.token`         | InfluxDB access token                                | String                                               | Yes      |
| `influx.bucket`        | InfluxDB bucket name                                 | String                                               | Yes      |



## Usage

### Spawning VMs
To spawn VMs, use the following command:
``` bash
spawner spawn -c <config-file-path>
```

### Destroying VMs
To destroy VMs, use the following command:
``` bash
spawner destroy -c <config-file-path>
```

### Listing VMs
To list VMs, use the following command:
``` bash
spawner list -c <config-file-path>
```