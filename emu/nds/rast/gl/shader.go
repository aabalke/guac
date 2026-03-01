package gl

import "github.com/aabalke/guac/emu/nds/utils"

type Shader struct {
	Texture *Texture
}

func NewShader() *Shader {
	return &Shader{}
}

const FACTOR = 32.0

func Modulate(vColor, tColor Color) Color {
	vColor.R *= tColor.R
	vColor.G *= tColor.G
	vColor.B *= tColor.B
	vColor.A *= tColor.A

	if vColor.R > 1 {
		vColor.R = 1
	}
	if vColor.G > 1 {
		vColor.G = 1
	}
	if vColor.B > 1 {
		vColor.B = 1
	}
	if vColor.A > 1 {
		vColor.A = 1
	}
	return vColor
}

func Decal(vColor, tColor Color) Color {
	at := tColor.A

	if at >= 1 {
		return tColor
	}

	vColor.R = tColor.R*at + vColor.R*(1-at)
	vColor.G = tColor.G*at + vColor.G*(1-at)
	vColor.B = tColor.B*at + vColor.B*(1-at)

	if vColor.R > 1 {
		vColor.R = 1
	}
	if vColor.G > 1 {
		vColor.G = 1
	}
	if vColor.B > 1 {
		vColor.B = 1
	}

	return vColor
}

var blendFunc = [...]func(texture *Texture, vColor, tColor Color) Color{
	func(texture *Texture, vColor, tColor Color) Color {
		return Modulate(vColor, tColor)

		con := func(iv, it float32) float32 {

			v := uint32(min(1, max(0, iv))*0x1F) << 1
			t := uint32(min(1, max(0, it))*0x1F) << 1

			out := ((t+1)*(v+1) - 1) >> 6

			return min(1, max(0, float32(out)/64))

			//v *= 0x1F
			//t *= 0x1F
			//return max(0, min(1, ((t)*(v)-1)/(FACTOR*FACTOR)))
		}

		vColor.R = con(vColor.R, tColor.R)
		vColor.G = con(vColor.G, tColor.G)
		vColor.B = con(vColor.B, tColor.B)
		vColor.A = con(vColor.A, tColor.A)
		return vColor
	},
	func(texture *Texture, vColor, tColor Color) Color {

		return Decal(vColor, tColor)

		//con := func(v, t, at float32) float32 {
		//	v *= FACTOR
		//	t *= FACTOR
		//	at *= FACTOR
		//	return max(0, min(1, (t*at+v*(FACTOR-at))/(FACTOR*FACTOR)))
		//}

		//vColor.R = con(vColor.R, tColor.R, tColor.A)
		//vColor.G = con(vColor.G, tColor.G, tColor.A)
		//vColor.B = con(vColor.B, tColor.B, tColor.A)
		//return vColor
	},
	func(texture *Texture, vColor, tColor Color) Color {

		if texture.IsHighlight {

			con := func(s, t float32) float32 {
				// assume s needs to be added as 0...1
				sb := s
				s *= FACTOR
				t *= FACTOR
				return max(0, min(1, ((t)*(s)-1)/(FACTOR*FACTOR)+sb))
			}

			toon := texture.ToonTbl[uint32(vColor.R*FACTOR)&0x1F]

			vColor.R = con(toon.R, tColor.R)
			vColor.G = con(toon.G, tColor.G)
			vColor.B = con(toon.B, tColor.B)
			vColor.A = con(vColor.A, tColor.A)
			return vColor
		}

		// toon

		con := func(s, t float32) float32 {
			s *= FACTOR
			t *= FACTOR
			return max(0, min(1, ((t)*(s)-1)/(FACTOR*FACTOR)))
		}

		toon := texture.ToonTbl[uint32(vColor.R*FACTOR)&0x1F]

		vColor.R = con(toon.R, tColor.R)
		vColor.G = con(toon.G, tColor.G)
		vColor.B = con(toon.B, tColor.B)
		vColor.A = con(vColor.A, tColor.A)
		return vColor
	},
	func(texture *Texture, vColor, tColor Color) Color {
		return Transparent
	},
}

func (s *Shader) Fragment(v *Vertex) {

	if s.Texture == nil {
		return
	}

	if s.Texture.CachedTexture != nil {
		tColor := s.Texture.Sample(v.Texture.X, v.Texture.Y)
		v.Color = blendFunc[s.Texture.Mode](s.Texture, v.Color, tColor)
		v.Color.A = utils.FloatRound(v.Color.A, 0.05)
		return
	}

	v.Color = blendFunc[s.Texture.Mode](s.Texture, v.Color, White)
	v.Color.A = utils.FloatRound(v.Color.A, 0.05)
}

func (s *Shader) SetTexture(texture *Texture) {
	s.Texture = texture
}
