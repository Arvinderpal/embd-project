package daemon

func (d *Daemon) GlobalStatus() (string, error) {
	logger.Infof("Received Staus Request...")
	return "Ok!", nil
}
