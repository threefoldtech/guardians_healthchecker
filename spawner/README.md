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

## Supported Configurations
Create a new configuration file, for example `config.yaml`

| Field              | Description                                          | Supported Values                                     |
| ------------------ | ---------------------------------------------------- | ---------------------------------------------------- |
| `farms`            | List of farm IDs where VMs should be deployed        | List of integers (e.g., `1`, `2`, etc.)              |
| `deployment_strategy` | Strategy for deploying VMs across nodes            | `"100%"`, `"70%"`, `"50%"`, etc.                     |
| `grid_endpoints`   | URLs for grid services                                |                                                      |
| `grid_endpoints.graphql`    | GraphQL endpoint URL                        | URL (e.g., `"https://graphql.dev.grid.tf/graphql"`)  |
| `grid_endpoints.proxy`      | Proxy endpoint URL                          | URL (e.g., `"https://gridproxy.dev.grid.tf/"`)       |
| `grid_endpoints.relay`      | Relay endpoint URL                          | URL (e.g., `"wss://relay.dev.grid.tf"`)              |
| `grid_endpoints.subsrate_url` | Substrate URL                             | URL (e.g., `"wss://tfchain.dev.grid.tf/ws"`)         |
| `failure_strategy` | Strategy for handling deployment failures            | `"retry"`, `"stop"`, `"destroy-all"`, `"destroy-failing"` |
| `mnemonic`         | Mnemonic for authentication                          | String                                               |
| `ssh_key`          | SSH key for accessing VMs                            | String                                               |
| `influx`           | InfluxDB configuration                               |                                                      |
| `influx.url`       | InfluxDB URL                                         | URL (e.g., `"http://influxdb.example.com"`)          |
| `influx.org`       | InfluxDB organization name                           | String                                               |
| `influx.token`     | InfluxDB access token                                | String                                               |
| `influx.bucket`    | InfluxDB bucket name                                 | String                                               |



## Usage

### Spawning VMs
To spawn VMs, use the following command:
``` bash
spawner spawn -c <config-file-path>
```

## Destroying VMs
To destroy VMs, use the following command:
``` bash
spawner destroy -c <config-file-path>
```

## Listing VMs
To list VMs, use the following command:
``` bash
spawner list -c <config-file-path>
```