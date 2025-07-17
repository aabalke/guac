package sdl

import "github.com/veandco/go-sdl2/sdl"

func InitPauseMenu(renderer *sdl.Renderer, scene *Scene) {

	c := sdl.Color{R: 228, G: 199, B: 153, A: 255}
	c2 := sdl.Color{R: 255, G: 255, B: 255, A: 255}

	l := NewLayout(&scene.H, &scene.W, 0, 0, 2)

	switch {
	case isGB:
		pause := NewGbMenu(scene, l, gbConsole, C_Brown)
		l = NewLayout(400, 200, 100, 100, 3)
		container := NewContainer(pause, l, C_Transparent, "evenlyVertical")

		text := "mute"
		if gbConsole.Muted {
			text = "unmute"
		}

		container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), "resume", 48, c, c2, ""))
		container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), text, 48, c, c2, ""))
		container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), "exit", 48, c, c2, ""))
		pause.Add(container)
		//pause.Add(NewText(s.Renderer, container, 5, "always save your game in the emulator before exiting", 16))
		scene.Add(pause)

		pause.InitOptions()
	case isGBA:
		pause := NewGbaMenu(scene, l, gbaConsole, C_Brown)
		l = NewLayout(400, 200, 100, 100, 3)
		container := NewContainer(pause, l, C_Transparent, "evenlyVertical")

		text := "mute"
		if gbaConsole.Muted {
			text = "unmute"
		}

		container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), "resume", 48, c, c2, ""))
		container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), text, 48, c, c2, ""))
		container.Add(NewText(container, NewLayout(0, 0, 0, 0, 5), "exit", 48, c, c2, ""))
		pause.Add(container)
		//pause.Add(NewText(s.Renderer, container, 5, "always save your game in the emulator before exiting", 16))
		scene.Add(pause)

		pause.InitOptions()
	}
}
