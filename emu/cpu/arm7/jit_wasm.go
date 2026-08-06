package arm7

func (j *Jit) CreateBlock(pc uint32, thumb bool)          {}
func (j *Jit) TestInstThumb(op uint16, f func(op uint16)) {}
func (j *Jit) TestInst(op uint32, f func(op uint32))      {}
