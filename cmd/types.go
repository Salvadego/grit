package cmd

const TaskDir = ".grit"
const StateFile = ".grit/.state.bin"
const StateMagic uint32 = 0x47726954
const StateVersion uint8 = 1

type Task struct {
	ID       uint64
	Status   string
	MTime    int64
	Title    string
	Tags     []string
	Filename string
}
