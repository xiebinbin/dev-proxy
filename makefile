
GOCMD = go
GOBUILD = $(GOCMD) build
TARGET = server

all: build

build:
	rm -f $(TARGET)
	$(GOBUILD) -o $(TARGET) main.go
clear:
	rm -f $(TARGET)
