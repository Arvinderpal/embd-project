				### Segue Start ###

sudo mkdir -p /var/run/segue ; sudo chmod 777 /var/run/segue
sudo ./segue --debug daemon run --n 192.168.80.201 -dr

				### Machine mh1 ###

sudo ./segue daemon machine join ../scripts/configs/mh-1.json
sudo ./segue daemon machine get mh1

				### Firmata Adaptor ###

sudo ./segue daemon adaptor attach ../scripts/configs/adaptor-serial-dev-ttyACM0.json
sudo ./segue daemon adaptor detach mh1 adaptor_firmata_serial ttyACM0

			### RaspberryPi Adaptor ###
sudo ./segue daemon adaptor attach ../scripts/configs/raspi/adaptor-raspi-mh1.json 

				### Unit Test ###

sudo ./segue daemon adaptor attach ../scripts/configs/adaptor-unittest.json
sudo ./segue daemon driver start ../scripts/configs/driver-unittest-mh1.json

				### Autonomous Car ###

sudo ./segue daemon machine join ../scripts/configs/mh-1.json; 
sudo ./segue daemon adaptor attach ../scripts/configs/adaptor-serial-dev-ttyACM0.json

sudo ./segue daemon driver start ../scripts/configs/dualmotors-mh1.json
sudo ./segue daemon driver start ../scripts/configs/ultrasonic-mh1.json; 
sudo ./segue daemon controller start ../scripts/configs/autonomous-drive-controller-mh1.json

				### RF24Network ###

sudo ./segue daemon controller start ../scripts/configs/raspi/rf24network-master-node-controller-mh1.json

				### Raspi LED Blink ###

sudo ./segue daemon driver start ../scripts/configs/raspi/led-mh1.json

					### GRPC ###

sudo ./segue daemon controller start ../scripts/configs/grpc-server-controller-mh1.json 


				### GORT ###
gort scan serial
gort arduino upload firmata /dev/ttyACM0
# Setup: to install avrdude: 
gort arduino install


				### socat ###
socat PTY,link=$HOME/dev/vmodem0,rawer,wait-slave \
EXEC:'"ssh modemserver.us.org socat - /dev/ttyS0,nonblock,rawer"'


sudo socat -d -d pty,link=/dev/ttyS0,raw,echo=0 pty,link=/dev/ttyACM0,raw,echo=0




######################## Manual Tests ############################

	# MPI -- testplugn
	sudo ./segue daemon machine join ../scripts/configs/mh-1.json
	sudo ./segue daemon controller start ../scripts/configs/raspi/mpi-controller-testplugin-mh1.json 

	# MPI -- rf24networknode
	sudo ./segue daemon machine join ../scripts/configs/mh-1.json
	sudo ./segue daemon controller start ../scripts/configs/raspi/mpi-controller-rf24network-master-mh1.json



			### RF24Network -- LED blink on master via child ###
	
	# MASTER:
		sudo ./segue daemon machine join ../scripts/configs/mh-1.json
		sudo ./segue daemon controller start ../scripts/configs/raspi/mpi-controller-rf24network-master-mh1.json
		sudo ./segue daemon adaptor attach ../scripts/configs/raspi/adaptor-raspi-mh1.json 
		sudo ./segue daemon driver start ../scripts/configs/raspi/led-mh1.json

	# SLAVE: 
		sudo ./segue daemon machine join ../scripts/configs/mh-1.json
		sudo ./segue daemon controller start ../scripts/configs/raspi/mpi-controller-rf24network-node-1-mh1.json
		sudo ./segue daemon adaptor attach ../scripts/configs/raspi/adaptor-raspi-mh1.json 
		sudo ./segue daemon controller start ../scripts/configs/grpc-server-controller-mh1.json
		sudo ./blink 


			### RFNetwork -- control car on master via child

	# MASTER:
		sudo ./segue daemon machine join ../scripts/configs/mh-1.json
		sudo ./segue daemon controller start ../scripts/configs/raspi/mpi-controller-rf24network-master-mh1.json
		sudo ./segue daemon adaptor attach ../scripts/configs/adaptor-serial-dev-ttyACM0.json
		sudo ./segue daemon driver start ../scripts/configs/dualmotors-mh1.json


	# SLAVE: 
		sudo ./segue daemon machine join ../scripts/configs/mh-1.json
		sudo ./segue daemon controller start ../scripts/configs/raspi/mpi-controller-rf24network-node-1-mh1.json		
		sudo ./segue daemon adaptor attach ../scripts/configs/raspi/adaptor-raspi-mh1.json 
		sudo ./segue daemon controller start ../scripts/configs/grpc-server-controller-mh1.json
		sudo ./remotecar 




