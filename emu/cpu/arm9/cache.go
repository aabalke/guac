package arm9

const (
	ADDRESS_SPACE = 0x1_0000_0000
	PAGE_SIZE     = ADDRESS_SPACE / 4096 // 4kb pages
)

type ProtectionUnit struct {
	Regions [8]uint32

	InstCache Cache
	DataCache Cache
	Enabled   bool
}

type Cache struct {
	Permissions uint32
	Ctrl        uint8
	Enabled     bool

	// bit mask 1 bit per page
	Pages [PAGE_SIZE / 64]uint64
}

func (p *ProtectionUnit) RefreshMPU() {
	clear(p.InstCache.Pages[:])
	clear(p.DataCache.Pages[:])

	if !p.Enabled {
		return
	}

	for i := range 8 {
		p.UpdateRegion(i)
	}
}

func (p *ProtectionUnit) UpdateRegion(n int) {
	if !p.Enabled {
		return
	}

	region := p.Regions[n]

	if enabled := region&1 != 0; !enabled {
		return
	}

	var (
		// uses 4kb sizes since minimum size of pu region size
		size  = max((region>>1)&0x1F, 11) - 11
		start = ((region >> 12) >> size) << size
		end   = start + (1 << size)
	)

	dbit := uint64(0)
	if p.DataCache.Enabled && (p.DataCache.Ctrl&(1<<n) != 0) {
		dbit = 1
	}

	ibit := uint64(0)
	if p.InstCache.Enabled && (p.InstCache.Ctrl&(1<<n) != 0) {
		ibit = 1
	}

	for i := start; i < end; i++ {
		idx := i / 64
		bit := i & 63
		p.InstCache.Pages[idx] = (p.InstCache.Pages[idx] &^ (1 << bit)) | (ibit << bit)
		p.DataCache.Pages[idx] = (p.DataCache.Pages[idx] &^ (1 << bit)) | (dbit << bit)
	}
}
