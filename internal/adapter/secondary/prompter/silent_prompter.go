package prompter

// SilentPrompter implements UserPrompter with no user interaction.
// Always returns default values.
type SilentPrompter struct{}

// NewSilentPrompter creates a new SilentPrompter instance.
func NewSilentPrompter() *SilentPrompter {
	return &SilentPrompter{}
}

// Confirm returns the default value.
func (p *SilentPrompter) Confirm(question string, defaultValue bool) bool {
	return defaultValue
}

// Select returns the default index.
func (p *SilentPrompter) Select(question string, options []string, defaultIndex int) int {
	return defaultIndex
}

// MultiSelect returns the default indices.
func (p *SilentPrompter) MultiSelect(question string, options []string, defaultIndices []int) []int {
	return defaultIndices
}

// Input returns the default value.
func (p *SilentPrompter) Input(question string, defaultValue string) string {
	return defaultValue
}

// Password returns an empty string.
func (p *SilentPrompter) Password(question string) string {
	return ""
}
