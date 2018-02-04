#include <stdio.h>
#include <stdlib.h>
#include <umad.h>
#include <ibnetdisc.h>

int main(int argc, char **argv) {
	char names[UMAD_MAX_DEVICES][UMAD_CA_NAME_LEN];
	int i, n;

	// umad_get_cas_names will not see the simulated sysfs directory from ibsim unless umad_init()
	// is called first.
	if (umad_init() < 0)
		IBPANIC("can't init UMAD library");

	if ((n = umad_get_cas_names(names, UMAD_MAX_DEVICES)) < 0)
		IBPANIC("can't list IB device names");

	for (i = 0; i < n; i++) {
		printf("%s\n", names[i]);
	}

	umad_done();
}
