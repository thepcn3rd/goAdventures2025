# STIX Indicator Parser

This Go program is designed to parse STIX (Structured Threat Information Expression) JSON files, extract specific indicators, and optionally send the extracted data to a Microsoft Teams webhook.


![Soldier Evaluating Threat Intel](/picts/scorpionSoldierThreatIntel.png)

## Usage

The provided prep.sh will download dependencies and compile the binary
```bash
./prep.sh
```

### Flags
```bash
./stixParser.bin -h
```

- `-f`: Specifies the STIX JSON file to load. Default is `default.json`.
- `-t`: Specifies the Teams webhook configuration file. Default is `teams.json`.

### Example Commands

1. **Basic Usage** (Download a stix file into the directory):
```bash
./stixParser.bin -f stix_data.json
```
   This command will load the `stix_data.json` file, parse the indicators, and print the extracted URLs to the console.

2. **Using a Custom Teams Webhook** (Place webhook in teams_config.json):
```bash
./stixParser.bin -f stix_data.json -t teamsSendMessage.json
```
   This command will load the `stix_data.json` file, parse the indicators, and send the first extracted URL to the Microsoft Teams channel specified in `teams_config.json`.


## Configuration

### Teams Webhook Configuration

The Teams webhook configuration is stored in a JSON file (default: `teams.json`). The file should contain the following structure:

```json
{
    "teamsWebhook": "YOUR_TEAMS_WEBHOOK_URL"
}
```

Replace `YOUR_TEAMS_WEBHOOK_URL` with the actual webhook URL provided by Microsoft Teams.

## Output

The program outputs the extracted URLs to the console. If a valid Teams webhook URL is provided, it will also send the first extracted URL to the specified Teams channel.

## Example STIX JSON File

Here is an example of what a STIX JSON file might look like:

```json
{
    "type": "bundle",
    "id": "bundle--12345678-1234-1234-1234-123456789012",
    "objects": [
        {
            "type": "indicator",
            "id": "indicator--12345678-1234-1234-1234-123456789012",
            "created": "2023-01-01T00:00:00Z",
            "modified": "2023-01-01T00:00:00Z",
            "pattern": "[url:value = 'http://example.com']",
            "valid_from": "2023-01-01T00:00:00Z",
            "pattern_type": "stix"
        }
    ]
}
```

## Dependencies

- Go standard library
- No external dependencies are required.

