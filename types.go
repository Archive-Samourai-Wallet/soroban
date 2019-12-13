package soroban

type Options struct {
	Redis OptionRedis
}

type OptionRedis struct {
	Hostname string
	Port     int
}
