set -e

# CGO_CFLAGS="-I/home/awander/go/src/github.com/raspberrypi/tools/arm-bcm2708/gcc-linaro-arm-linux-gnueabihf-raspbian-x64/bin/../arm-linux-gnueabihf/libc/usr/local/include" 

CGO_CXXFLAGS="-Wall -fPIC -Ofast -mfpu=vfp -mfloat-abi=hard -mtune=arm1176jzf-s -std=c++0x -I/home/awander/go/src/github.com/raspberrypi/tools/arm-bcm2708/gcc-linaro-arm-linux-gnueabihf-raspbian-x64/bin/../arm-linux-gnueabihf/libc/usr/local/include" CGO_LDFLAGS="-L/home/awander/go/src/github.com/Arvinderpal/embd-project/playground/rf24-swig/tmrh20/rf24 -lrf24 -L/home/awander/go/src/github.com/Arvinderpal/embd-project/playground/rf24-swig/tmrh20/rf24 -lrf24-bcm -L/home/awander/go/src/github.com/Arvinderpal/RF24Network -lrf24network" CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc  CXX=arm-linux-gnueabihf-g++  GOOS=linux GOARCH=arm  GOARM=7  go build -x

scp helloworld pi@raspberrypi:/home/pi/embd-project/segue
scp helloworld pi@pi2b1:/home/pi/embd-project/segue
