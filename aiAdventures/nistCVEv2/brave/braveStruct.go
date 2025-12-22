package brave

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type BraveConfiguration struct {
	BraveURL        string            `json:"braveURL"`
	SearchKeyword   string            `json:"searchKeyword,omitempty"`
	FullURL         string            `json:"fullURL,omitempty"`
	BraveAPIKey     string            `json:"braveAPIKey"`
	ResultCount     int               `json:"resultCount"`     // number of results to return
	Freshness       string            `json:"freshness"`       // pd past day, pw past week, pm past month, py past year
	SafeSearch      string            `json:"safeSearch"`      // off, moderate, strict
	TextDecorations string            `json:"textDecorations"` // true, false
	Summary         string            `json:"summary"`         // true, false
	RequestHeaders  map[string]string `json:"requestHeaders"`
}

func (b *BraveConfiguration) BuildFullURL() error {
	values := url.Values{}
	values.Add("q", b.SearchKeyword)
	values.Add("count", fmt.Sprintf("%d", b.ResultCount))
	values.Add("freshness", b.Freshness)
	values.Add("safeSearch", b.SafeSearch)
	values.Add("textDecorations", b.TextDecorations)
	values.Add("summary", b.Summary)

	b.FullURL = b.BraveURL + "?" + values.Encode()

	return nil
}

func (b *BraveConfiguration) SubmitRequest() (SearchResults, error) {
	err := b.BuildFullURL()
	if err != nil {
		return SearchResults{}, err
	}

	//log.Println("Brave Search URL: " + b.FullURL + "\n")

	req, err := http.NewRequest("GET", b.FullURL, nil)
	if err != nil {
		return SearchResults{}, err
	}

	// Add the Request Headers
	for key, value := range b.RequestHeaders {
		// Take the Brave API from the configuration and populate the token value
		if key == "X-Subscription-Token" {
			value = b.BraveAPIKey
		}
		//log.Printf("Adding Header: %s: %s\n", key, value)
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return SearchResults{}, fmt.Errorf("unable to pull the response %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SearchResults{}, fmt.Errorf("unable to read the response body %v", err)
	}

	//log.Printf("Response Body\n%s\n\n", string(body))

	// Parse Brave API response (structure would depend on their API)
	var braveResponse SearchResults
	if err := json.Unmarshal(body, &braveResponse); err != nil {
		return SearchResults{}, fmt.Errorf("unable to unmarshal the brave response")
	}

	return braveResponse, nil
}

type SearchResults struct {
	Mixed MixedSection `json:"mixed"`
	Query QueryInfo    `json:"query"`
	Type  string       `json:"type"`
	Web   WebResults   `json:"web"`
}

type MixedSection struct {
	Main []MixedItem `json:"main"`
	Side []MixedItem `json:"side"`
	Top  []MixedItem `json:"top"`
	Type string      `json:"type"`
}

type MixedItem struct {
	All   bool   `json:"all"`
	Index int    `json:"index"`
	Type  string `json:"type"`
}

type QueryInfo struct {
	BadResults           bool   `json:"bad_results"`
	City                 string `json:"city"`
	Country              string `json:"country"`
	HeaderCountry        string `json:"header_country"`
	IsNavigational       bool   `json:"is_navigational"`
	IsNewsBreaking       bool   `json:"is_news_breaking"`
	MoreResultsAvailable bool   `json:"more_results_available"`
	Original             string `json:"original"`
	PostalCode           string `json:"postal_code"`
	ShouldFallback       bool   `json:"should_fallback"`
	ShowStrictWarning    bool   `json:"show_strict_warning"`
	SpellcheckOff        bool   `json:"spellcheck_off"`
	State                string `json:"state"`
}

type WebResults struct {
	FamilyFriendly bool        `json:"family_friendly"`
	Results        []WebResult `json:"results"`
	Type           string      `json:"type"`
}

type WebResult struct {
	Age            string    `json:"age"`
	Description    string    `json:"description"`
	FamilyFriendly bool      `json:"family_friendly"`
	IsLive         bool      `json:"is_live"`
	IsSourceBoth   bool      `json:"is_source_both"`
	IsSourceLocal  bool      `json:"is_source_local"`
	Language       string    `json:"language"`
	MetaURL        MetaURL   `json:"meta_url"`
	PageAge        string    `json:"page_age"`
	Profile        Profile   `json:"profile"`
	Subtype        string    `json:"subtype"`
	Thumbnail      Thumbnail `json:"thumbnail"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	URL            string    `json:"url"`
}

type MetaURL struct {
	Favicon  string `json:"favicon"`
	Hostname string `json:"hostname"`
	Netloc   string `json:"netloc"`
	Path     string `json:"path"`
	Scheme   string `json:"scheme"`
}

type Profile struct {
	Img      string `json:"img"`
	LongName string `json:"long_name"`
	Name     string `json:"name"`
	URL      string `json:"url"`
}

type Thumbnail struct {
	Logo     bool   `json:"logo"`
	Original string `json:"original"`
	Src      string `json:"src"`
}
