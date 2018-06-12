package sprite

type State struct{}

func (s *State) Start() {}

func (s *State) OnSize(w, h float32) {}

func (s *State) Ctl(itf interface{}) {}

func (s *State) OnFrame(ind, vert []byte, next uint16) ([]byte, []byte, uint16, bool) {}
