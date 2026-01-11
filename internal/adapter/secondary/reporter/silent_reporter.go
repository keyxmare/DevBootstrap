package reporter

// SilentReporter implements ProgressReporter with no output.
type SilentReporter struct{}

// NewSilentReporter creates a new SilentReporter instance.
func NewSilentReporter() *SilentReporter {
	return &SilentReporter{}
}

// Info does nothing.
func (r *SilentReporter) Info(message string) {}

// Success does nothing.
func (r *SilentReporter) Success(message string) {}

// Warning does nothing.
func (r *SilentReporter) Warning(message string) {}

// Error does nothing.
func (r *SilentReporter) Error(message string) {}

// Progress does nothing.
func (r *SilentReporter) Progress(message string) {}

// ClearProgress does nothing.
func (r *SilentReporter) ClearProgress() {}

// Section does nothing.
func (r *SilentReporter) Section(title string) {}

// Step does nothing.
func (r *SilentReporter) Step(current, total int, message string) {}

// Header does nothing.
func (r *SilentReporter) Header(title string) {}

// Summary does nothing.
func (r *SilentReporter) Summary(title string, items map[string]string) {}
