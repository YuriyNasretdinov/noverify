package lintdebug

var (
	callbacks []func(string)
)

// Send a debug message
func Send(msg string, args ...interface{}) {

}

// Register debug events receiver. There must be only one receiver
func Register(cb func(string)) {
	callbacks = append(callbacks, cb)
}
