GOOS=linux GOARCH=arm  GOARM=7  go build -x 
scp testplugin pi@raspberrypi:/var/run/segue/plugins
