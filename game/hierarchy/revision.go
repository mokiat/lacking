package hierarchy

const initialRevision int32 = -1

var freeRevision int32

func nextRevision() int32 {
	freeRevision++
	return freeRevision
}
