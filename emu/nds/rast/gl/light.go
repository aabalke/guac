package gl

type LightData struct {
    Lights [4]Light
    Normal Vector
    DiffuseColor Color
    AmbientColor Color
    SpecularColor Color
    EmissionColor Color
    UseSpecularTbl bool
    ShininessTbl ShininessTbl
}

type Light struct {
    Vector Vector
    HalfVector Vector
    Color Color
}

type ShininessTbl [32 * 8]float32
