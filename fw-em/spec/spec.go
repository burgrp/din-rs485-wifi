package spec

type Parity uint8

const (
	ParityEven Parity = iota
	ParityOdd
	ParityNone
)

type WordOrder uint8

const (
	WordOrderHighFirst WordOrder = iota
	WordOrderLowFirst
)

type Config struct {
	MeterAddress uint8
	Parity       Parity
	WordOrder    WordOrder
	Baud         uint32
	PollMs       uint32
}

const (
	RegVoltage1           = 0x0000 + 1
	RegVoltage2           = 0x0002 + 1
	RegVoltage3           = 0x0004 + 1
	RegCurrent1           = 0x0008 + 1
	RegCurrent2           = 0x000A + 1
	RegCurrent3           = 0x000C + 1
	RegPowerActiveTotal   = 0x0010 + 1
	RegPowerActive1       = 0x0012 + 1
	RegPowerActive2       = 0x0014 + 1
	RegPowerActive3       = 0x0016 + 1
	RegPowerReactiveTotal = 0x0018 + 1
	RegPowerReactive1     = 0x001A + 1
	RegPowerReactive2     = 0x001C + 1
	RegPowerReactive3     = 0x001E + 1
	RegPowerFactor1       = 0x002A + 1
	RegPowerFactor2       = 0x002C + 1
	RegPowerFactor3       = 0x002E + 1
	RegFrequency          = 0x0036 + 1
	RegEnergyActive       = 0x0100 + 1
	RegEnergyReactive     = 0x0400 + 1
)

type Run struct {
	Base  uint16
	First uint8
	Count uint8
}

var PollRuns = [...]Run{
	{Base: 0x0000, First: 0, Count: 3},
	{Base: 0x0008, First: 3, Count: 3},
	{Base: 0x0010, First: 6, Count: 8},
	{Base: 0x002A, First: 14, Count: 3},
	{Base: 0x0036, First: 17, Count: 1},
	{Base: 0x0100, First: 18, Count: 1},
	{Base: 0x0400, First: 19, Count: 1},
}
