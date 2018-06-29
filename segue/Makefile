include ../Makefile.defs

TARGET=segue
BASE=$(GOPATH)/src/github.com/Arvinderpal/$(TARGET)
GOVENDOR=govendor
GO=go

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

all: $(TARGET)

run:
	./segue

clean:
	go clean
	rm -f $(TARGET)
	rm -f $(TARGET_ARM)
	# rm -f ../common/matrapb/*.pb.go 

testall:
	$Q ../testing/unit/test_all.sh
# go test ../pkg/...
# go test ../daemon/...


vendor: #$(info $(M) retrieving dependencies…)
	$Q cd $(BASE) && $(GOVENDOR) add +external 


SOURCES := $(shell find ../daemon ../common ../pkg . -name '*.go')
TEST_LIST := $(shell find ../daemon ../common ../pkg . -name '*_test.go')
# TEST_LIST = $(foreach int, $(SOURCES), $(int)_test)


$(TARGET): $(SOURCES)
	protoc -I=../common/seguepb --go_out=plugins=grpc:../common/seguepb ../common/seguepb/*.proto
	# protoc -I=../common/seguepb --gofast_out=../common/seguepb ../common/seguepb/*.proto
	
	go build -race -ldflags "-X "github.com/Arvinderpal/common".Version=$(VERSION)" -o $(TARGET) ./main.go
	@touch main.go


arm: 
	protoc -I=../common/seguepb --go_out=plugins=grpc:../common/seguepb ../common/seguepb/*.proto
	GOARM=7 GOARCH=arm GOOS=linux go build -ldflags "-X "github.com/Arvinderpal/common".Version=$(VERSION)" -o $(TARGET) ./main.go
	
	# CGO_CXXFLAGS="-Wall -fPIC -Ofast -mfpu=vfp -mfloat-abi=hard -mtune=arm1176jzf-s -std=c++0x -I/home/awander/go/src/github.com/raspberrypi/tools/arm-bcm2708/gcc-linaro-arm-linux-gnueabihf-raspbian-x64/bin/../arm-linux-gnueabihf/libc/usr/local/include" CGO_LDFLAGS="-L/home/awander/go/src/github.com/Arvinderpal/embd-project/playground/rf24-swig/tmrh20/rf24 -lrf24 -L/home/awander/go/src/github.com/Arvinderpal/embd-project/playground/rf24-swig/tmrh20/rf24 -lrf24-bcm " CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc  CXX=arm-linux-gnueabihf-g++  GOOS=linux GOARCH=arm  GOARM=7  go build -x 
	@touch main.go


