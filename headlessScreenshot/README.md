# Headless Browser Screenshot Tool

This tool is designed to capture screenshots of web pages using a headless browser. It supports multiple URLs and user agents, allowing for flexible and automated screenshot generation. The tool also extracts URLs from the HTML content of the pages and saves them alongside the screenshots and HTML files.

## Features

- **Multiple URL Support**: Load URLs from a configuration file or directly from a list.
- **User Agent Rotation**: Use different user agents for each screenshot.
- **HTML and URL Extraction**: Save the HTML content of the page and extract all URLs found in the HTML.
- **Customizable Browser Settings**: Configure the browser's viewport size and page load time.

![Scorpion Soldier in Factory](/picts/scorpionSoldierFactory.png)
## Installation

A prep script is included in the directory of this tool.  You do need to change the path in the tool to be valid.  Then it will set the paths, download the dependencies and compile the tool.

```bash
./prep.sh
```

## Configuration

The tool uses a JSON configuration file (`config.json`) to specify the URLs, user agents, and browser settings. Below is an example configuration:

```json
{
  "urlOption": "File",
  "_urlOptionNotes": "The above option can be 'File' or 'List'",
  "urlListFile": "urls.txt",
  "urlList": [],
  "outputPathPNG": "screenshots",
  "headlessBrowserConfig": {
    "widthPNG": 1920,
    "heightPNG": 1080,
    "timeToLoadPageSeconds": 5
  },
  "userAgents": [
    {
      "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
      "description": "Chrome on Windows 10"
    },
    {
      "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
      "description": "Safari on macOS"
    }
  ]
}
```

### Configuration Fields

- **urlOption**: Specifies whether to load URLs from a file or from the `urlList` field.
- **urlListFile**: Path to a file containing a list of URLs (one per line).
- **urlList**: A list of URLs with descriptions (used if `urlOption` is not `"File"`).
- **outputPathPNG**: Directory where screenshots, HTML files, and URL lists will be saved.
- **headlessBrowserConfig**: Configuration for the headless browser, including viewport size and page load time.
- **userAgents**: A list of user agents to use for capturing screenshots.

## Usage

1. **Prepare the Configuration**: Edit the `config.json` file to include the URLs you want to capture and the user agents you want to use.

2. **Run the Tool**:
   ```bash
   ./headlessScreenshot.bin -config config.json
   ```

3. **Output**: The tool will generate screenshots, HTML files, and URL lists in the specified output directory. Each file is named with a timestamp, URL count, and user agent count for easy identification.


## Acknowledgments

- [chromedp](https://github.com/chromedp/chromedp) for providing the headless browser automation.
