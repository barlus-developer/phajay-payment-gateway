package phajay

type Phajay struct {
	key string
}

func New(key string) *Phajay {
	return &Phajay{
		key: key,
	}
}