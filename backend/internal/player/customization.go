package player

// DefaultCosmetics is used if a join request omits customization fields.
import "github.com/nir-game/server/internal/protocol"

func DefaultCosmetics() protocol.Cosmetics {
	return protocol.Cosmetics{
		ColorPrimary:   "#3b82f6",
		ColorSecondary: "#1e3a8a",
		HatID:          "none",
	}
}

// ValidCosmeticOptions lists what the customization UI is allowed to pick
// from. Keep this in sync with frontend/src/player/customization.ts.
//
// TODO(player owner): replace with real cosmetic catalog (skins, hats),
// possibly tied into economy/ for unlockable cosmetics.
var ValidHatIDs = []string{"none", "cap", "helmet", "beanie"}

// ValidateJoin fills in sane defaults for any missing/invalid cosmetic
// fields on a join request. Called by the server before spawning a player.
func ValidateJoin(data *protocol.JoinData) {
	if data.Name == "" {
		data.Name = "Player"
	}
	if data.ColorPrimary == "" {
		data.ColorPrimary = DefaultCosmetics().ColorPrimary
	}
	if data.ColorSecondary == "" {
		data.ColorSecondary = DefaultCosmetics().ColorSecondary
	}
	if data.HatID == "" {
		data.HatID = "none"
	}
}
