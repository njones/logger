module github.com/njones/logger

go 1.13

require (
	github.com/njones/logger/color v0.0.0
	github.com/njones/logger/kv v0.0.0
)

replace (
	github.com/njones/logger/color => ./color
	github.com/njones/logger/kv => ./kv
)
