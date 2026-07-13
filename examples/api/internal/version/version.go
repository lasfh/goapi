package version

import "time"

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = ""
)

func BuildTimeOrNow() string {
	if BuildTime == "" {
		return time.Now().Format(time.RFC3339)
	}

	return BuildTime
}
