# API Keys Reference

This file documents API keys used during development. **These keys are for personal test accounts with minimal rate limits.** Do not use in production.

## Key Status (as of January 2026)

| Service | Key Present | Rate Limit | Status |
|---------|-------------|------------|--------|
| OpenAQ | ✅ | 1000 req/day | Free tier |
| AlphaVantage | ✅ | 25 req/day | Free tier |
| Grid.dev | ✅ | 100 req/day | Free tier |
| NASDAQ | ✅ | 50 req/day | Free tier |
| FRED | ✅ | Unlimited | Public API |
| Ember | ✅ | 1000 req/day | Free tier |
| USDA NASS | ✅ | Unlimited | Public API |
| EIA | ✅ | 5000 req/day | Free tier |
| Movebank | ✅ | N/A | Academic account |

## Security Note

- **Archived Project:** These keys remain in `.env` for reference but pose minimal risk (test accounts, expired, or public-tier)
- **For Active Projects:** Use environment variables, AWS Secrets Manager, or HashiCorp Vault
- **Never Commit:** Always add `.env` to `.gitignore` (already configured in this repo)

## How Keys Were Used

```go
// Example: go-ingest/internal/clients/openaq.go
apiKey := os.Getenv("OPENAQ_API_KEY")
req.Header.Set("X-API-Key", apiKey)
```

## Replacement Instructions (If Needed)

1. Visit service provider (e.g., https://api.openaq.org/)
2. Register for free account
3. Copy API key to `go-ingest/.env`
4. Restart Go API server

---

**Archive Note:** This file is for documentation only. The project is no longer actively developed.
