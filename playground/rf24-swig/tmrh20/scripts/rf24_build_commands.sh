

CGO_LDFLAGS="-L/home/awander/go/src/github.com/Arvinderpal/embd-project/playground/rf24-swig/tmrh20/rf24 -lrf24" CGO_CFLAGS="-I. -I/home/awander/go/src/github.com/Arvinderpal/embd-project/playground/rf24-swig/tmrh20/rf24" CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc  CXX=arm-linux-gnueabihf-g++  GOOS=linux GOARCH=arm  GOARM=6  go build -x



CGO_LDFLAGS="-L/home/awander/go/src/github.com/Arvinderpal/embd-project/playground/rf24-swig/tmrh20/rf24 -lrf24 " CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc  CXX=arm-linux-gnueabihf-g++  GOOS=linux GOARCH=arm  GOARM=7  go build -x


CGO_LDFLAGS="-L/home/awander/go/src/github.com/Arvinderpal/embd-project/playground/rf24-swig/tmrh20/rf24 -lrf24 -Wl,-rpath -Wl,\$ORIGIN/../rf24" CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc  CXX=arm-linux-gnueabihf-g++  GOOS=linux GOARCH=arm  GOARM=6  go build -x




