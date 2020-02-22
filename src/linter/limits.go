package linter

// ParseWaiter waits to allow parsing of a file.
type ParseWaiter struct {
	size int
}

// MemoryLimiterThread starts memory limiter goroutine that disallows to use parse files more than MaxFileSize
// total bytes.
func MemoryLimiterThread() {
}

// BeforeParse must be called before parsing file, so that soft memory
// limit can be applied.
// Do not forget to call Finish()!
func BeforeParse(size int, filename string) *ParseWaiter {
	return &ParseWaiter{
		size: size,
	}
}

// Finish must be called after parsing is finished (e.g. using defer p.Finish()) to
// allow other goroutines to parse files.
func (p *ParseWaiter) Finish() {

}
