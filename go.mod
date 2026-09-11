module github.com/aabalke/guac

// local development
//replace github.com/hajimehoshi/dialog => C:\dev\repos\emulators\dialog
//replace github.com/ebitenui/ebitenui => C:\dev\repos\emulators\ebitenui
//replace github.com/aabalke/gojit => C:\dev\repos\jit\gojit

// release version
replace github.com/ebitenui/ebitenui => github.com/aabalke/ebitenui v0.0.0-20260507040224-7e5cd031ea7d

replace github.com/hajimehoshi/dialog => github.com/aabalke/dialog v0.0.0-20260806052813-02b04fc6c149

// gojit, ebitenui, and dialog will need to be updated to proper versioning
// before release. unversioned packaged are v0.0.0-YYYYMMDD______-CCCCCCCCCCCC
// where C is the beginning of the commit

go 1.27.1

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/ebitenui/ebitenui v0.7.3
	github.com/hajimehoshi/dialog v0.0.0-20260703050910-dfca0e7cf198
	github.com/hajimehoshi/ebiten/v2 v2.10.1
)

require (
	github.com/ebitengine/gomobile v0.0.0-20260820040257-d11f821a26a6 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/oto/v3 v3.5.0 // indirect
	github.com/ebitengine/purego v0.11.0 // indirect
	github.com/frustra/bbcode v0.0.0-20201127003707-6ef347fbe1c8 // indirect
	github.com/go-text/typesetting v0.3.5 // indirect
	github.com/jfreymuth/pulse v0.1.3 // indirect
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/image v0.45.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
