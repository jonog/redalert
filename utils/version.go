package utils

var (
	version string
	build   string
)

func RegisterVersionAndBuild(v, b string) {
	version = v
	build = b
}

func Version() string {
	return version
}

func Build() string {
	return build
}
