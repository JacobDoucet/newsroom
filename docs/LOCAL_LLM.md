# Local LLM (No-cost Dev)

The backend talks to an OpenAI-compatible Chat Completions API.

Recommended: LM Studio on macOS, running an OpenAI-compatible local server.

Default config (Docker containers call host):
- LLM_BASE_URL=http://host.docker.internal:1234/v1
- LLM_API_KEY=local-dev
- LLM_MODEL=<your-local-model-name>

## LM Studio
1. Install LM Studio
2. Download an instruct model suitable for JSON/structured output
3. Start the local server (OpenAI compatible)
4. Set LLM_MODEL to the exact model identifier used by LM Studio

## llama.cpp alternative
Run llama-server on host and expose /v1/chat/completions.
Set LLM_BASE_URL accordingly.

## Notes
- Model output must be valid JSON. The backend retries once if parsing fails.
- Store raw model output for debugging.