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

		ResultChan   chan Result
		ProgressChan chan Progress
	}

	Progress struct {
		Total    int64
		Done     int64
		Progress float64
	}

	Result struct {
		Target   string
		Port     int
		Protocol string
		User     string
		Pass     string
	}
)
