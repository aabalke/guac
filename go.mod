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

go 1.26.5

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/aabalke/gojit v0.0.0-20260616021404-5808a49d96fc
	github.com/ebitenui/ebitenui v0.7.3
	github.com/hajimehoshi/dialog v0.0.0-20260703050910-dfca0e7cf198
	github.com/hajimehoshi/ebiten/v2 v2.9.7
	golang.org/x/sys v0.46.0
)

require (
	github.com/ebitengine/gomobile v0.0.0-20250923094054-ea854a63cce1 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/oto/v3 v3.4.0 // indirect
	github.com/ebitengine/purego v0.10.1 // indirect
	github.com/edsrzf/mmap-go v1.2.0 // indirect
	github.com/frustra/bbcode v0.0.0-20201127003707-6ef347fbe1c8 // indirect
	github.com/go-text/typesetting v0.3.0 // indirect
	github.com/jezek/xgb v1.1.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/image v0.31.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)
