ROOT_DIR=../..

include $(ROOT_DIR)/Makefile.include

FILES=$(wildcard *.cc)
MODULES=$(FILES:%.cc=%.o)

INCLUDE=-I$(ROOT_DIR)/engine/src/Network/protobuf-2.4.1/src

TARGET=librpc.so
DEPEND_LIB=-lprotobuf

$(TARGET):$(MODULES)
	@$(CC) $(BUILD_SYSTEM) $(SHARED_LIB) $(TARGET) $(MODULES) -L$(ROOT_DIR)/$(OUTPUT_DIR) $(DEPEND_LIB);
	@echo Linking ..... $(TARGET)
	@cp $(TARGET) $(ROOT_DIR)/$(OUTPUT_DIR)
	@echo "Copy " $(TARGET) " to " $(ROOT_DIR)/$(OUTPUT_DIR)

$(MODULES):%.o:%.cc
	@$(CC) $(BUILD_SYSTEM) $(MODULE) $(INCLUDE) $< -o $@
	@echo Compiling ..... $<
clean :
	rm $(TARGET) *.o
