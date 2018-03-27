
#include <alsa/asoundlib.h>
#include <signal.h>
#include <stdlib.h>

extern void availWriteCallback(snd_async_handler_t *ahandler);

static inline int cSetAvailWriteCallback (snd_pcm_t *handle) {
	snd_async_handler_t *ahandler;
	struct sigaction sa;
	int err;

	err = snd_async_add_pcm_handler(&ahandler, handle, availWriteCallback, NULL);
	if (err < 0) {
		fprintf(stderr, "Audio: can't add availWriteCallback. E: %s\n", snd_strerror(err));
		return -1;
	}
	sigaction(SIGIO, NULL, &sa);
	sa.sa_flags |= SA_ONSTACK;
	sigaction(SIGIO, &sa, NULL);
	return 0;
}
