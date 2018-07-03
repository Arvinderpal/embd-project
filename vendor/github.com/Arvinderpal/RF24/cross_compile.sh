export PATH=$PATH:/home/awander/go/src/github.com/raspberrypi/tools/arm-bcm2708/gcc-linaro-arm-linux-gnueabihf-raspbian-x64/bin 
make
ln -s librf24.so.1.3.1 librf24.so
ln -s librf24.so librf24-bcm.so
