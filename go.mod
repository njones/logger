module github.com/njones/logger

go 1.13

require (
	github.com/njones/logger/color v1.0.3
	github.com/njones/logger/kv v1.0.3
)

replace (
	github.com/njones/logger/color => ./color
	github.com/njones/logger/kv => ./kv
)
