package apu

const (
	CPU_FREQ_HZ              = 16777216
	SND_FREQUENCY            = 32768 // sample rate
	SND_SAMPLES              = 512
	SAMP_CYCLES              = (CPU_FREQ_HZ / SND_FREQUENCY)
	BUFF_SAMPLES             = ((SND_SAMPLES) * 16 * 2)
	BUFF_SAMPLES_MSK         = ((BUFF_SAMPLES) - 1)
	SAMPLE_TIME      float64 = 1.0 / SND_FREQUENCY
	STREAM_LEN               = (2 * 2 * SND_FREQUENCY / 60) - (2*2*SND_FREQUENCY/60)%4
)
