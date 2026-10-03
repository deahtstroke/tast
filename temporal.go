package tast

type TemporalUnit int

const (
	Years TemporalUnit = iota + 1
	Months
	Days
	Hours
	Minutes
	Seconds
)

func (t TemporalUnit) String() string {
	switch t {
	case Years:
		return "Years"
	case Months:
		return "Months"
	case Days:
		return "Days"
	case Hours:
		return "Hours"
	case Minutes:
		return "Minutes"
	case Seconds:
		return "Seconds"
	default:
		return ""
	}
}
