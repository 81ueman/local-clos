package update

type PathSegmentType uint8

const (
	AS_SET      PathSegmentType = 1
	AS_SEQUENCE PathSegmentType = 2
)

type PathSegmentLength uint8
type AS uint16

type ASPathSegment struct {
	PathSegmentType   PathSegmentType
	PathSegmentLength PathSegmentLength
	AS                []AS
}
