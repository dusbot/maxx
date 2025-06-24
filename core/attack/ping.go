package attack

type ping struct {
}

func (p *ping) Name() string {
	return "[ATK-Ping]"
}

func (p *ping) Attack(in Input) (err error) {
	// if task := in.GetTask(); task != nil {
	// 	for _, target := range task.Targets {
	// 		if ipv4, err0 := utils.IsValidIP(target); err0 != nil {
	// 			if task.Verbose {
	// 				slog.Printf(slog.WARN, "%s: %s is not a valid IP address ,skip", p.Name(), target)
	// 			}
	// 			continue
	// 		} else {
	// 			pingStatus, err1 := ping_.Ping(target, ping_.PingOptions{
	// 				Count:   1,
	// 				Timeout: time.Duration(task.Timeout),
	// 				IsIPv6:  !ipv4,
	// 			})
	// 			if err1 != nil {
	// 				if task.Verbose {
	// 					slog.Printf(slog.DEBUG, "%s: %s ping failed: %v", p.Name(), target, err1)
	// 				}
	// 				continue
	// 			}
	// 		}

	// 	}
	// }
	return
}
