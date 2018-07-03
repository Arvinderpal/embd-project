OS=LINUX
SOC=BCM2835
DRIVER=RPi
CPUFLAGS=-march=armv7-a -mtune=arm1176jzf-s -mfpu=vfp -mfloat-abi=hard
CFLAGS=-march=armv7-a -mtune=arm1176jzf-s -mfpu=vfp -mfloat-abi=hard -Ofast -Wall -pthread 
PREFIX=/usr/local
REMOTE_PREFIX=/usr/local
LIB=rf24
LIBNAME=librf24.so.1.3.1
LIB_VERSION=1.3.1
LIBSYMLINKS=librf24.so.1.3 librf24.so.1 librf24.so
LIBDEPRECATE=librf24-bcm.so
CC=arm-linux-gnueabihf-gcc
CXX=arm-linux-gnueabihf-g++
LIB_DIR=/usr/local/lib
REMOTE_LIB_DIR=/usr/local/lib
HEADER_DIR=/usr/local/include/RF24
REMOTE_HEADER_DIR=/usr/local/include/RF24
DRIVER_DIR=utility/RPi
ARCH_DIR=utility
REMOTE=pi@raspberrypi
REMOTE_HOST=raspberrypi
REMOTE_PORT=22
SHARED_LINKER_FLAGS= -pthread -shared -Wl,-soname,librf24.so.1
LDCONFIG=ldconfig
REMOTE_LDCONFIG=/sbin/ldconfig
EXAMPLES_DIR=/usr/local/bin
REMOTE_EXAMPLES_DIR=/usr/local/bin
