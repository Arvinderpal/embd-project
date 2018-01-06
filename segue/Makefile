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
	# rm -f ../common/matrapb/*.pb.go 


SOURCES := $(shell find ../daemon ../common ../pkg . -name '*.go')
TEST_LIST := $(shell find ../daemon ../common ../pkg . -name '*_test.go')
# TEST_LIST = $(foreach int, $(SOURCES), $(int)_test)


$(TARGET): $(SOURCES)
	# protoc -I=../common/matrapb --go_out=../common/matrapb ../common/matrapb/*.proto
	# protoc -I=../common/matrapb --gofast_out=../common/matrapb ../common/matrapb/*.proto
	
	go build -race -ldflags "-X "github.com/Arvinderpal/common".Version=$(VERSION)" -o $(TARGET) ./main.go
	@touch main.go

testall:
	$Q ../testing/unit/test_all.sh
# go test ../pkg/...
# go test ../daemon/...


vendor: #$(info $(M) retrieving dependencies…)
	$Q cd $(BASE) && $(GOVENDOR) add +external 

