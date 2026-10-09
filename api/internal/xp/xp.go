package xp

type Bucket string

const (
	Under30m Bucket = "under-30m"
	M30to60  Bucket = "30-60m"
	OneTo2h  Bucket = "1-2h"
	HalfDay  Bucket = "half-day"
	FullDay  Bucket = "full-day"
)

const (
	CompletionBonus = 20
	PhotoBonus      = 30
)

var Buckets = []Bucket{Under30m, M30to60, OneTo2h, HalfDay, FullDay}

var base = map[Bucket]int{
	Under30m: 10,
	M30to60:  20,
	OneTo2h:  40,
	HalfDay:  80,
	FullDay:  120,
}

func BaseXP(b Bucket) int {
	return base[b]
}

func ValidBucket(s string) bool {
	_, ok := base[Bucket(s)]
	return ok
}

func Award(b Bucket, hasPhoto bool) int {
	x := BaseXP(b) + CompletionBonus
	if hasPhoto {
		x += PhotoBonus
	}
	return x
}

type Level struct {
	Level  int
	Into   int
	Needed int
}

func LevelForXp(total int) Level {
	level := 0
	for 50*(level+1)*(level+2) <= total {
		level++
	}
	return Level{
		Level:  level,
		Into:   total - 50*level*(level+1),
		Needed: 100 * (level + 1),
	}
}

func Title(level int) string {
	titles := []string{
		"Grass Toucher",
		"Sidequester",
		"Wanderer",
		"Trailblazer",
		"Pathfinder",
		"Wayfarer",
		"Legend",
	}
	if level < len(titles) {
		return titles[level]
	}
	return titles[len(titles)-1]
}
