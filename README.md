# agent-platform by A2A-go

## Configuration

The agent platform uses a TOML configuration file to manage agents. Create a `config.toml` file in the project root with the following structure:

```toml
[[agents]]
name = "agent_name"
url = "${AGENT_URL_1}"

[[agents]]
name = "another_agent"
url = "http://localhost:8081"
```

### Environment Variables

You can use environment variables in your configuration by wrapping them in `${}` syntax. The platform will automatically replace these with the actual environment variable values.

### Example Configuration

See `config.example.toml` for a complete example of how to configure multiple agents.

### Fallback

If the configuration file is not found, the platform will fall back to using environment variables:
- `AIR_AGENT_URL`
- `WEA_AGENT_URL`