// Providers PicoClaw talks to without an API key: local inference servers
// that serve an OpenAI-compatible API with no authentication by default. The
// server keeps the same list (server/service/picoclaw/model_config.go).
const KEYLESS_PROVIDERS = new Set(['lmstudio', 'ollama', 'vllm']);

// Whether a provider/model identifier such as "ollama/qwen3.5:9b" names a
// provider that works without an API key.
export function modelProviderAllowsEmptyApiKey(modelIdentifier: string): boolean {
  const trimmed = modelIdentifier.trim();
  const slash = trimmed.indexOf('/');
  if (slash <= 0) {
    return false;
  }
  return KEYLESS_PROVIDERS.has(trimmed.slice(0, slash).trim().toLowerCase());
}
