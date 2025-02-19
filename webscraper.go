package webscraper

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/agent-api/core"
	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/core/types"
	"github.com/agent-api/gsv"

	"github.com/PuerkitoBio/goquery"
)

// WebScraperAgent extends the DefaultAgent with web scraping capabilities
type WebScraperAgent struct {
	*agent.DefaultAgent

	config *WebScraperConfig
}

type WebScraperConfig struct {
	Provider  core.Provider
	MaxSteps  int
	Logger    *slog.Logger
	UserAgent string
}

const (
	FetchWebPageTool   = "fetch_webpage"
	ExtractContentTool = "extract_content"
)

type fetchWebPageParams struct {
	URL *gsv.StringSchema `json:"url"`
}

type extractContentParams struct {
	HTML *gsv.StringSchema `json:"html"`
}

// NewWebScraperAgent creates a new web scraper agent with necessary tools
func NewWebScraperAgent(config *WebScraperConfig) (*WebScraperAgent, error) {
	if config.UserAgent == "" {
		config.UserAgent = "WebScraperAgent/1.0"
	}

	baseAgent := agent.NewAgent(&agent.NewAgentConfig{
		Provider: config.Provider,
		MaxSteps: config.MaxSteps,
		Logger:   config.Logger,
	})

	scraper := &WebScraperAgent{
		DefaultAgent: baseAgent,
		config:       config,
	}

	// Add web scraping specific tools
	if err := scraper.initializeTools(); err != nil {
		return nil, fmt.Errorf("failed to initialize tools: %w", err)
	}

	return scraper, nil
}

func (w *WebScraperAgent) initializeTools() error {
	// Fetch web page tool
	fetchSchema := &fetchWebPageParams{
		URL: gsv.String().Description("The URL to fetch"),
	}
	compiledFetch, err := gsv.CompileSchema(fetchSchema, &gsv.CompileSchemaOpts{
		SchemaTitle: "FetchWebPage",
	})
	if err != nil {
		return err
	}

	fetchTool, err := types.WrapToolFunction(w.fetchWebPage)
	if err != nil {
		return err
	}

	if err := w.AddTool(types.Tool{
		Name:                FetchWebPageTool,
		Description:         "Fetches the content of a webpage",
		WrappedToolFunction: fetchTool,
		JSONSchema:          compiledFetch,
	}); err != nil {
		return err
	}

	// Extract content tool
	extractContentSchema := &extractContentParams{
		HTML: gsv.String().Description(""),
	}
	compiledExtractContentSchema, err := gsv.CompileSchema(extractContentSchema, &gsv.CompileSchemaOpts{
		SchemaTitle: "extractContentTextTool",
	})
	if err != nil {
		return err
	}

	extractContentTool, err := types.WrapToolFunction(w.extractContent)
	if err != nil {
		return err
	}

	if err := w.AddTool(types.Tool{
		Name:                ExtractContentTool,
		Description:         "",
		WrappedToolFunction: extractContentTool,
		JSONSchema:          compiledExtractContentSchema,
	}); err != nil {
		return err
	}

	return nil
}

// Tool implementations
func (w *WebScraperAgent) fetchWebPage(ctx context.Context, args *fetchWebPageParams) (interface{}, error) {
	url, ok := args.URL.Value()
	if !ok {
		return nil, fmt.Errorf("missing required URL parameter")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", w.config.UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch webpage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch webpage: status code %d", resp.StatusCode)
	}

	var content strings.Builder
	_, err = io.Copy(&content, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return content.String(), nil
}

func (w *WebScraperAgent) extractContent(ctx context.Context, args *extractContentParams) (interface{}, error) {
	htmlContent, _ := args.HTML.Value()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Remove unwanted elements
	doc.Find("script, style").Remove()

	// Extract main content
	var content strings.Builder
	doc.Find("body").Each(func(i int, s *goquery.Selection) {
		content.WriteString(s.Text())
	})

	return strings.TrimSpace(content.String()), nil
}
