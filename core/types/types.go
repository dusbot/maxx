package types

type (
	Task struct {
		Verbose   bool
		MaxTime   int
		Timeout   int
		Interval  int
		NoUser    bool
		Progress  bool
		Thread    int
		Targets   []string
		Users     []string
		Passwords []string
	}

	Result struct {
		Target   string
		Port     int
		Protocol string
		User     string
		Pass     string
	}
)
