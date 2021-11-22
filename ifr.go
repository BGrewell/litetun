package litetun

type Ifr struct {
	Name [16]byte
	Flags uint16
	_ [22]byte
}
