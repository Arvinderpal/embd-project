set -e

PROJECT_HOME="/home/awander/go/src/github.com/Arvinderpal/embd-project"

# Build binaries:
cd $PROJECT_HOME/segue; make arm

# SCP binaries and config files:
# pi@raspberrypi
scp $PROJECT_HOME/segue/segue pi@raspberrypi:/home/pi/embd-project/segue
scp -r $PROJECT_HOME/scripts/configs pi@raspberrypi:/home/pi/embd-project/scripts
# pi@pi2b1
# scp $PROJECT_HOME/segue/segue pi@pi2b1:/home/pi/embd-project/segue
# scp -r $PROJECT_HOME/scripts/configs pi@pi2b1:/home/pi/embd-project/scripts


### Additional Binaries Go Here ###

## REMOTECAR ##
# cd $PROJECT_HOME/playground/seguepb-clients/remotecar; make arm
# scp $PROJECT_HOME/playground/seguepb-clients/remotecar/remotecar pi@raspberrypi:/home/pi/embd-project/segue
# scp $PROJECT_HOME/playground/seguepb-clients/remotecar/remotecar pi@pi2b1:/home/pi/embd-project/segue

## BLINK ##
# cd $PROJECT_HOME/playground/seguepb-clients/blink; make arm
# scp $PROJECT_HOME/playground/seguepb-clients/blink/blink pi@raspberrypi:/home/pi/embd-project/segue
# scp $PROJECT_HOME/playground/seguepb-clients/blink/blink pi@pi2b1:/home/pi/embd-project/segue
