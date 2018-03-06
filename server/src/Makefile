ROOT_DIR=../..

include $(ROOT_DIR)/Makefile.include

MODULES_DIR=fightserver tlogserver tpserver Robot

ARCH:=$(shell uname -s)
#只在linux下编译tlog
ifneq ($(ARCH),Linux)
        MODULES_DIR=fightserver Robot
endif

all:
	@for i in $(MODULES_DIR); do \
		cd $$i &&make && cd ..;\
	done;

clean:
	@for i in $(MODULES_DIR); do\
		cd $$i && make clean && cd ..;\
	done;
