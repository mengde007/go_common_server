ROOT_DIR=.

include $(ROOT_DIR)/Makefile.include

MODULES_DIR=engine rpc client server bin

all:
	@for i in $(MODULES_DIR); do \
		cd $$i &&make && cd ..;\
	done;

clean:
	@for i in $(MODULES_DIR); do\
		echo $$i && cd $$i && make clean && cd ..;\
	done;
