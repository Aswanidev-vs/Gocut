package filters

func BuildLoudNormFilter() string {
	return "loudnorm=I=-16:TP=-1.5:LRA=11"
}

func BuildNoiseReductionFilter() string {
	return "afftdn=nf=-25"
}
