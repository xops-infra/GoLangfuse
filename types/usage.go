package types

// UsageUnit a model usage unit
type UsageUnit string

// Characters LLM usage unit in characters, this is used by google
// Tokens LLM usage unit in tokens, this is used by OpenAI and bedrock
// Milliseconds LLM usage unit in millis
// Seconds LLM usage unit in seconds
// Images LLM usage unit for image
const (
	Characters   UsageUnit = "CHARACTERS"
	Tokens       UsageUnit = "TOKENS"
	Milliseconds UsageUnit = "MILLISECONDS"
	Seconds      UsageUnit = "SECONDS"
	Images       UsageUnit = "IMAGES"
)

// Usage represents token and cost usage for language model interactions in Langfuse.
// It tracks input, output, and total usage, as well as associated costs and token breakdowns.
// Fields:
//   - Input, Output, Total: Quantities measured in the specified unit (characters, tokens, etc.).
//   - Unit: The unit of measurement (see UsageUnit).
//   - InputCost, OutputCost, TotalCost: Cost values for input, output, and total usage.
//   - PromptTokens, CompletionTokens, TotalTokens: Token counts for prompt, completion, and total (for LLMs).
type Usage struct {
	Input            int       `json:"input,omitempty"`
	Output           int       `json:"output,omitempty"`
	Total            int       `json:"total,omitempty"`
	Unit             UsageUnit `json:"unit,omitempty"`
	InputCost        float64   `json:"inputCost,omitempty"`
	OutputCost       float64   `json:"outputCost,omitempty"`
	TotalCost        float64   `json:"totalCost,omitempty"`
	PromptTokens     int       `json:"promptTokens,omitempty"`
	CompletionTokens int       `json:"completionTokens,omitempty"`
	TotalTokens      int       `json:"totalTokens,omitempty"`
}
