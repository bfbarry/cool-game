package combat

// Weapon stats. Cost is in Points and should match any corresponding
// economy/store.go StoreItem with the same ID.
type Weapon struct {
	ID          string
	Name        string
	Damage      int
	FireRateMs  int // minimum ms between shots
	BulletSpeed float64
	Range       float64
	Cost        int
	Tier        int
}

// Catalog is the authoritative weapon list. Starter weapon ("pistol") is
// free and given to every new player; the rest are store-purchasable
// upgrades (see economy/store.go).
//
// TODO(combat owner): tune numbers; add more tiers.
var Catalog = map[string]Weapon{
	"pistol": {ID: "pistol", Name: "Pistol", Damage: 10, FireRateMs: 400, BulletSpeed: 600, Range: 500, Cost: 0, Tier: 0},
	"smg":    {ID: "smg", Name: "SMG", Damage: 8, FireRateMs: 120, BulletSpeed: 700, Range: 450, Cost: 300, Tier: 1},
	"rifle":  {ID: "rifle", Name: "Rifle", Damage: 22, FireRateMs: 250, BulletSpeed: 900, Range: 700, Cost: 700, Tier: 2},
	"sniper": {ID: "sniper", Name: "Sniper", Damage: 60, FireRateMs: 900, BulletSpeed: 1400, Range: 1200, Cost: 1500, Tier: 3},
}

func WeaponByID(id string) (Weapon, bool) {
	w, ok := Catalog[id]
	return w, ok
}
