CFLAGS := -I/usr/include/infiniband
LDFLAGS := -libmad -libumad -libnetdisc

all:
	gcc $(CFLAGS) -o netdisc netdisc.c $(LDFLAGS)
