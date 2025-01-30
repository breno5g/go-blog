package config

var (
	logger *Logger
	paths  Paths
)

func GetLogger(prefix string) *Logger {
	logger = NewLogger(prefix)
	return logger
}

func GetPaths() Paths {
	paths = InitilizeConstants()
	return paths
}
