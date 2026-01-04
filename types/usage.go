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
	Input            int       `json:"input,omitempty" valid:"range(0|9999999)"`
	Output           int       `json:"output,omitempty" valid:"range(0|9999999)"`
	Total            int       `json:"total,omitempty" valid:"range(0|9999999)"`
	Unit             UsageUnit `json:"unit,omitempty" valid:"-"`
	InputCost        float64   `json:"inputCost,omitempty" valid:"range(0|999999)"`
	OutputCost       float64   `json:"outputCost,omitempty" valid:"range(0|999999)"`
	TotalCost        float64   `json:"totalCost,omitempty" valid:"range(0|999999)"`
	PromptTokens     int       `json:"promptTokens,omitempty" valid:"range(0|9999999)"`
	CompletionTokens int       `json:"completionTokens,omitempty" valid:"range(0|9999999)"`
	TotalTokens      int       `json:"totalTokens,omitempty" valid:"range(0|9999999)"`
}

// UsageBuilder provides a fluent interface for building Usage
type UsageBuilder struct {
	usage *Usage
}

// NewUsage creates a new UsageBuilder
func NewUsage() *UsageBuilder {
	return &UsageBuilder{
		usage: &Usage{
			Unit: Tokens, // Default to tokens
		},
	}
}

// WithTokens sets token counts
func (b *UsageBuilder) WithTokens(input, output int) *UsageBuilder {
	b.usage.Input = input
	b.usage.Output = output
	b.usage.Total = input + output
	b.usage.PromptTokens = input
	b.usage.CompletionTokens = output
	b.usage.TotalTokens = input + output
	b.usage.Unit = Tokens
	return b
}

// WithCosts sets cost values
func (b *UsageBuilder) WithCosts(inputCost, outputCost float64) *UsageBuilder {
	b.usage.InputCost = inputCost
	b.usage.OutputCost = outputCost
	b.usage.TotalCost = inputCost + outputCost
	return b
}

// WithUnit sets the usage unit
func (b *UsageBuilder) WithUnit(unit UsageUnit) *UsageBuilder {
	b.usage.Unit = unit
	return b
}

// WithCharacters sets character counts
func (b *UsageBuilder) WithCharacters(input, output int) *UsageBuilder {
	b.usage.Input = input
	b.usage.Output = output
	b.usage.Total = input + output
	b.usage.Unit = Characters
	return b
}

// Build returns the built Usage
func (b *UsageBuilder) Build() Usage {
	return *b.usage
}

type UsageDetail struct {
	Input                    int `json:"input,omitempty" valid:"range(0|9999999)"`
	Output                   int `json:"output,omitempty" valid:"range(0|9999999)"`
	Image                    int `json:"image,omitempty" valid:"range(0|9999999)"`
	OutputReasoning          int `json:"output_reasoning,omitempty" valid:"-"`
	Total                    int `json:"total,omitempty" valid:"range(0|9999999)"`
	InputCacheRead           int `json:"input_cache_read,omitempty" valid:"range(0|9999999)"`
	InputCachedTokens        int `json:"input_cached_tokens,omitempty" valid:"range(0|9999999)"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty" valid:"range(0|9999999)"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty" valid:"range(0|9999999)"`
}

type CostDetail struct {
	Input                    float64 `json:"input,omitempty"`
	Output                   float64 `json:"output,omitempty"`
	Total                    float64 `json:"total,omitempty"`
	Image                    float64 `json:"image,omitempty"`
	InputCachedTokens        float64 `json:"input_cached_tokens,omitempty"`
	CacheCreationInputTokens float64 `json:"cache_creation_input_tokens,omitempty"`
	OutputReasoning          float64 `json:"output_reasoning,omitempty"`
}
