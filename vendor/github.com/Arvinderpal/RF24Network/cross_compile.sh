export PATH=$PATH:/home/awander/go/src/github.com/raspberrypi/tools/arm-bcm2708/gcc-linaro-arm-linux-gnueabihf-raspbian-x64/bin 

# PREFIX=/usr/local
PREFIX=/home/awander/go/src/github.com/raspberrypi/tools/arm-bcm2708/gcc-linaro-arm-linux-gnueabihf-raspbian-x64/bin/../arm-linux-gnueabihf/libc/usr/local

# Library parameters
# where to put the lib
LIBDIR=$PREFIX/lib
# lib name 
LIB_RFN=librf24network
# shared library name
LIBNAME_RFN=$LIB_RFN.so.1.0

HEADER_DIR=$PREFIX/include/RF24Network

ARCH=armv7-a

CCFLAGS="-Ofast -mfpu=vfp -mfloat-abi=hard -march=$ARCH -mtune=arm1176jzf-s -std=c++0x -I/home/awander/go/src/github.com/raspberrypi/tools/arm-bcm2708/gcc-linaro-arm-linux-gnueabihf-raspbian-x64/bin/../arm-linux-gnueabihf/libc/usr/local/include "

LLFLAGS="-L/home/awander/go/src/github.com/raspberrypi/tools/arm-bcm2708/gcc-linaro-arm-linux-gnueabihf-raspbian-x64/bin/../arm-linux-gnueabihf/libc/usr/local/lib -lrf24-bcm"

arm-linux-gnueabihf-g++ -Wall -fPIC $CCFLAGS -c RF24Network.cpp

arm-linux-gnueabihf-g++ -shared -Wl,-soname,librf24network.so.1 $CCFLAGS -o $LIBNAME_RFN RF24Network.o $LLFLAGS

ln -s librf24network.so.1.0 librf24network.so
