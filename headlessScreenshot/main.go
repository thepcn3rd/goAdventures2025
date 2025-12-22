package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

type Configuration struct {
	URLOption     string              `json:"urlOption"`
	URLListFile   string              `json:"urlListFile"`
	URLList       []URLListStruct     `json:"urlList"`
	OutputPath    string              `json:"outputPathPNG"`
	BrowserConfig BrowserConfigStruct `json:"headlessBrowserConfig"`
	UserAgents    []UserAgentStruct   `json:"userAgents"`
}

type URLListStruct struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type BrowserConfigStruct struct {
	WidthPNG   int `json:"widthPNG"`
	HeightPNG  int `json:"heightPNG"`
	TimeToLoad int `json:"timeToLoadPageSeconds"`
}

type UserAgentStruct struct {
	UserAgent   string `json:"userAgent"`
	Description string `json:"description"`
}

func loadConfig(cPtr string) Configuration {

	var c Configuration
	fmt.Println("Loading the following config file: " + cPtr + "\n")
	// go logToSyslog(fmt.Sprintf("Loading the following config file: %s\n", *ConfigPtr))
	configFile, err := os.Open(cPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	// var config Configuration
	if err := decoder.Decode(&c); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	return c
}

func captureScreenshot(url, outputPath, outputPathHTML, outputPathURL, userAgent string, c Configuration) error {
	// Create a new context for ChromeDP
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, (time.Duration(c.BrowserConfig.TimeToLoad)*2)*time.Second)
	defer cancel()

	// Allocate buffer for the screenshot
	var buf []byte
	var htmlContent string

	// Run tasks to navigate to the page and take a screenshot
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(c.BrowserConfig.HeightPNG), int64(c.BrowserConfig.WidthPNG)),
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`navigator.__defineGetter__('userAgent', function(){ return '%s'; });`, userAgent), nil).Do(ctx)
		}),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(time.Duration(c.BrowserConfig.TimeToLoad)*time.Second),
		chromedp.WaitReady(`body`, chromedp.ByQuery),
		chromedp.FullScreenshot(&buf, 100),
		chromedp.OuterHTML(`html`, &htmlContent),
	)
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Save PNG File
	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	// Save HTML File
	if err := os.WriteFile(outputPathHTML, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to save html: %w", err)
	}

	// Extract URLs from HTML File
	urls := extractURLs(htmlContent)
	urlsContent := []byte{}
	for _, url := range urls {
		urlsContent = append(urlsContent, []byte(url+"\n")...)
	}

	// Save extracted URLs to a file
	if err := os.WriteFile(outputPathURL, urlsContent, 0644); err != nil {
		return fmt.Errorf("failed to save extracted URLs: %w", err)
	}

	return nil
}

func extractURLs(htmlContent string) []string {
	urlRegex := regexp.MustCompile(`https?://[^"'>]+`)
	return urlRegex.FindAllString(htmlContent, -1)
}

func createURLListFromFile(filename string) []URLListStruct {
	var urlList []URLListStruct
	var urlStruct URLListStruct

	file, err := os.Open(filename)
	cf.CheckError("Unable to open the file specified in config as urlListFile", err, true)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parsedURL, err := url.Parse(line)
		if err != nil {
			fmt.Printf("Skipping invalid URL: %s\n", line)
			continue
		}
		urlStruct.URL = parsedURL.String()
		urlList = append(urlList, urlStruct)

	}

	if err := scanner.Err(); err != nil {
		cf.CheckError("Unable to scan the file specified as urlListFile", err, true)
	}

	return urlList
}

func main() {
	var config Configuration
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	config = loadConfig(*ConfigPtr)

	// Identify the URLs that are involved in the browsing with the headless browser
	if config.URLOption == "File" {
		config.URLList = createURLListFromFile(config.URLListFile)
	}

	cf.CreateDirectory("/" + config.OutputPath)

	// Loop through the URLs
	for urlCount, url := range config.URLList {
		// Loop through the User Agents Specified
		for uaCount, userAgent := range config.UserAgents {
			// Filename contains the urlCount and uaCount
			timestamp := time.Now().Format("2006-01-02_15-04")
			urlString := strings.Replace(url.URL, "http://", "", 1)
			urlString = strings.Replace(urlString, "https://", "", 1)
			urlString = strings.ReplaceAll(urlString, ".", "_")

			pngFilename := config.OutputPath + "/screenshot_" + urlString + "_" + timestamp + "_" + strconv.Itoa(urlCount) + "_" + strconv.Itoa(uaCount) + ".png"
			htmlFilename := config.OutputPath + "/html_" + urlString + "_" + timestamp + "_" + strconv.Itoa(urlCount) + "_" + strconv.Itoa(uaCount) + ".txt"
			urlFilename := config.OutputPath + "/url_" + urlString + "_" + timestamp + "_" + strconv.Itoa(urlCount) + "_" + strconv.Itoa(uaCount) + ".txt"
			fmt.Printf("Capturing screenshot for the URL: %s with User Agent: %s\n", url.URL, userAgent.UserAgent)
			if err := captureScreenshot(url.URL, pngFilename, htmlFilename, urlFilename, userAgent.UserAgent, config); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Printf("Screenshot saved to %s\n", pngFilename)
			fmt.Printf("HTML saved to %s\n", htmlFilename)
			fmt.Printf("URL saved to %s\n\n", urlFilename)
		}
	}
}
