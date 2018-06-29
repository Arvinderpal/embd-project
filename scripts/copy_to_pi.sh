set -e

					### REMOTE HOSTS ###
declare -a HOSTS=(	"raspberrypi" 
                	"pi2b1"
)

PROJECT_HOME="/home/awander/go/src/github.com/Arvinderpal/embd-project"

					### Segue ###
# Build
cd $PROJECT_HOME/segue; make arm

# SCP binaries and config files:
for i in "${HOSTS[@]}"
do
	echo ">>>>>>>>> Uploading to $i <<<<<<<<<<<"
	scp $PROJECT_HOME/segue/segue pi@$i:/home/pi/embd-project/segue
	scp -r $PROJECT_HOME/scripts/configs pi@$i:/home/pi/embd-project/scripts
done


					### Plugins ###
# RF24Network Plugin
cd $PROJECT_HOME/plugins/rf24networknode; ./build.sh
for i in "${HOSTS[@]}"
do
	echo ">>>>>>>>> Uploading rf24networknode to $i <<<<<<<<<<<"
	scp rf24networknode pi@$i:/etc/segue/plugins
done

					### Clients ###

	## REMOTECAR ##
cd $PROJECT_HOME/playground/seguepb-clients/remotecar; make arm
for i in "${HOSTS[@]}"
do
	echo ">>>>>>>>> Uploading remotecar to $i <<<<<<<<<<<"
	scp $PROJECT_HOME/playground/seguepb-clients/remotecar/remotecar pi@$i:/home/pi/embd-project/segue
done

	## BLINK ##
cd $PROJECT_HOME/playground/seguepb-clients/blink; make arm
for i in "${HOSTS[@]}"
do
	echo ">>>>>>>>> Uploading blink to $i <<<<<<<<<<<"
	scp $PROJECT_HOME/playground/seguepb-clients/blink/blink pi@$i:/home/pi/embd-project/segue
done

