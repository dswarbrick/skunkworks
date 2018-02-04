#include <stdio.h>
#include <stdlib.h>
#include <umad.h>
#include <ibnetdisc.h>

void main() {
	int ret;

	ret = umad_init();
	printf("umad_init(): %d\n", ret);

	umad_done();
}
