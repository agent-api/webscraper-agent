# agent-api Web scraper agent

This module provides a simple, off the shelf web scraper AI agent with tools for
querying content off the web and extracting the relevant HTML content.

## Usage

Install the module:

```
go get github.com/agent-api/webscraper-agent
```

Use the agent with your specified provider:

```go
provider := openai.NewProvider(&openai.ProviderOpts{
	Logger: logger,
})
provider.UseModel(ctx, gpt4o.GPT4_O)


scraper, _ := webscraper.NewWebScraperAgent(&webscraper.WebScraperConfig{
	Provider: provider,
	Logger:   logger,
})

result, err := scraper.Run(ctx, "Please scrape https://johncodes.com/archive/2025/01-11-whats-an-ai-agent/ and summarize it.", agent.DefaultStopCondition)
if err != nil {
	panic(err)
}
```
