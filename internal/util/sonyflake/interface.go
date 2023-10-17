package sonyflake

type IdGenerator interface {
	GenerateId() (uint64, error)
}
