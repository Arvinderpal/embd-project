sudo mkdir -p /var/run/segue ; sudo chmod 777 /var/run/segue
./segue --debug daemon run --n 192.168.80.201 -dr

./segue daemon machine join ../scripts/configs/mh-1.json ;./segue daemon adaptor attach ../scripts/configs/adaptor-serial-dev-ttyACM0.json
./segue daemon machine get mh1

./segue daemon adaptor detach mh1 adaptor_firmata_serial ttyACM0


# Drivers:
./segue daemon driver start ../scripts/configs/unittest-mh1.json
./segue daemon driver start ../scripts/configs/dualmotors-mh1.json

./segue daemon driver stop mh1 driver_dualmotors dualmotors-1

#Controllers:
./segue daemon controller start ../scripts/configs/



 socat PTY,link=$HOME/dev/vmodem0,rawer,wait-slave \
EXEC:'"ssh modemserver.us.org socat - /dev/ttyS0,nonblock,rawer"'


sudo socat -d -d pty,link=/dev/ttyS0,raw,echo=0 pty,link=/dev/ttyACM0,raw,echo=0


./segue daemon machine join ../scripts/configs/mh-1.json ;./segue daemon adaptor attach ../scripts/configs/adaptor-serial-dev-ttyACM0.json

./segue daemon driver start ../scripts/configs/dualmotors-mh1.json; ./segue daemon driver start ../scripts/configs/ultrasonic-mh1.json; ./segue daemon controller start ../scripts/configs/autonomous-drive-controller-mh1.json


### RF24Network ###
./segue daemon controller start ../scripts/configs/raspi/rf24network-master-node-controller-mh1.json

# GORT
gort scan serial
gort arduino upload firmata /dev/ttyACM0
# Setup: to install avrdude: 
gort arduino install
