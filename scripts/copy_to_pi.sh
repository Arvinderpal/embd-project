PROJECT_HOME="/home/awander/go/src/github.com/Arvinderpal/embd-project"
cd $PROJECT_HOME/segue; make arm
cd $PROJECT_HOME/playground/seguepb-clients/remotecar; make arm
scp $PROJECT_HOME/segue/segue pi@raspberrypi:/home/pi/embd-project/segue
scp $PROJECT_HOME/playground/seguepb-clients/remotecar/remotecar pi@raspberrypi:/home/pi/embd-project/segue
scp -r $PROJECT_HOME/scripts/configs pi@raspberrypi:/home/pi/embd-project/scripts

