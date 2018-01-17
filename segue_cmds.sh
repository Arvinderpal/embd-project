
./segue daemon machine join ../scripts/configs/mh-1.json ;./segue daemon adaptor attach ../scripts/configs/adaptor-serial-dev-ttyACM0.json
./segue daemon driver start ../scripts/configs/dualmotors-mh1.json
./segue daemon machine get mh1

./segue daemon adaptor detach mh1 adaptor_firmata_serial ttyACM0
./segue daemon driver stop mh1 driver_dualmotors dualmotors-1



 socat PTY,link=$HOME/dev/vmodem0,rawer,wait-slave \
EXEC:'"ssh modemserver.us.org socat - /dev/ttyS0,nonblock,rawer"'


sudo socat -d -d pty,link=/dev/ttyS0,raw,echo=0 pty,link=/dev/ttyACM0,raw,echo=0


