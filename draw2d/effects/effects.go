package effects

type Effect interface {
	Start()
	OnFrame([]byte, []byte, uint16) ([]byte, []byte, uint16, bool)
	OnSize(float32, float32)
	Ctl(interface{})
}
