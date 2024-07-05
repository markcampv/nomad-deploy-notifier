# Nomad Deploy Notifier

Nomad Deploy Notifier is a tool to send Nomad deployment messages to InfluxDB and Splunk. This tool allows you to monitor Nomad deployments, node events, and job events by sending these events to your chosen monitoring and logging platforms.

## Description

Nomad Deploy Notifier subscribes to Nomad event streams and sends relevant event data to InfluxDB and/or Splunk based on user-defined criteria. It supports filtering events by specific topics (e.g., Deployment, Node, Job) and job names.

## Available Flags

- `-influxdb`: Enable sending data to InfluxDB.
- `-splunk`: Enable sending data to Splunk HEC.
- `-topics`: Comma-separated list of topics to send to Splunk. Valid topics are `Deployment`, `Node`, and `Job`. This flag is case-insensitive.
- `-job_name`: Name of the job to filter events. Only events related to this job will be sent.

## Environment Variables

Ensure the following environment variables are set when running the tool:

For InfluxDB:
- `INFLUXDB_TOKEN`: Your InfluxDB token.
- `INFLUXDB_URL`: Your InfluxDB URL.
- `INFLUXDB_ORG`: Your InfluxDB organization.
- `INFLUXDB_BUCKET`: Your InfluxDB bucket.

For Splunk:
- `SPLUNK_HEC_TOKEN`: [Your Splunk HEC token](https://kinneygroup.com/blog/http-event-collector/).
- `SPLUNK_HEC_ENDPOINT`: Your Splunk HEC endpoint URL. 
  - Splunk Cloud Ex: https://prd-p-g0qdhx.splunkcloud.com:8088/services/collector/event
  - Splunk Enterprise Ex: https://localhost:8088/services/collector/event

## Usage Examples

### Send Deployment Messages to InfluxDB

```sh
./nomad-deploy-notifier -influxdb -topics=deployment,node -job_name=quick_test
```

### Send Deployment Messages to Splunk
```sh
./nomad-deploy-notifier -splunk -topics=job,node -job_name=quick_test
```

### Installation

You can download pre-built binaries of Nomad Deploy Notifier directly from our GitHub [Releases](https://github.com/markcampv/nomad-deploy-notifier/releases) page. This allows you to install the tool without needing to compile it from the source.

1. **Go to the Releases Page**: Navigate to [Releases](https://github.com/markcampv/nomad-deploy-notifier/releases) in the Nomad Deploy Notifier repository.

2. **Download the Binary**: Download the appropriate binary for your operating system and architecture. We provide binaries for Windows, macOS, and Linux.

    - For Linux: `nomad-deploy-notifier-linux-amd64.zip`
    - For macOS: `nomad-deploy-notifier-darwin-amd64.zip`
    - For Windows: `nomad-deploy-notifier-windows-amd64.zip`

3. **Make the Binary Executable** (Linux and macOS):

   After downloading, you may need to make the binary executable. On Linux and macOS, you can do this with the following command:

```sh
   chmod +x nomad-deploy-notifier 

