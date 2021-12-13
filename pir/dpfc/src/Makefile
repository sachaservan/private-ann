TARGET = test
CFLAGS = -O3
LDFLAGS = -lcrypto -lssl -lm

$(TARGET): test.o libdpf.a
	$(CC) $^ -o $@ $(LDFLAGS)

test.o: test.c ../include/dpf.h
	gcc $(CFLAGS) -c $< -o $@ $(LDFLAGS)

libdpf.a: dpf.o 
	ar rcs $@ $^

dpf.o: dpf.c ../include/dpf.h
	gcc $(CFLAGS) -c -o $@ $< $(LDFLAGS)

clean:
	rm -f *.o *.a $(TARGET)
