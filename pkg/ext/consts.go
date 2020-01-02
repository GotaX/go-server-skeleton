package ext

import "os"

const (
	EnvAppVersion = "APP_VERSION"
)

func HostName() string {
	if host, err := os.Hostname(); err == nil {
		return host
	} else if host, ok := os.LookupEnv("HOSTNAME"); ok {
		return host
	} else {
		return "unknown"
	}
}

func HostIPStr() string {
	if ip, err := HostIP(); err != nil {
		return "unknown"
	} else {
		return ip.String()
	}
}

func Version() string {
	if version, ok := os.LookupEnv(EnvAppVersion); ok {
		return version
	} else {
		return "dev"
	}
}

func SetVersion(version string) {
	_ = os.Setenv(EnvAppVersion, version)
}
