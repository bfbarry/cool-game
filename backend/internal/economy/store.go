package economy

import "github.com/nir-game/server/internal/protocol"

// Catalog is the authoritative list of purchasable items. Weapon costs
// should match combat.Catalog; vehicle costs should match vehicles.Catalog.
// Duplicated here deliberately (rather than importing combat/vehicles) so
// economy/ stays a leaf package — see CONTEXT.md module table.
//
// TODO(economy owner): consider generating this from combat/vehicles
// catalogs at startup instead of hand-duplicating costs, once those
// packages stabilize.
var Catalog = []protocol.StoreItem{
	{ID: "smg", Name: "SMG", Kind: "weapon", Cost: 300},
	{ID: "rifle", Name: "Rifle", Kind: "weapon", Cost: 700},
	{ID: "sniper", Name: "Sniper", Kind: "weapon", Cost: 1500},
	{ID: "helicopter", Name: "Helicopter", Kind: "vehicle", Cost: 5000},
}

func ItemByID(id string) (protocol.StoreItem, bool) {
	for _, item := range Catalog {
		if item.ID == id {
			return item, true
		}
	}
	return protocol.StoreItem{}, false
}

// Spender is the subset of Player behavior Purchase needs: read current
// points, deduct them, and learn the result. Keeps economy/ from importing
// player/.
type Spender interface {
	PointsBalance() int
	DeductPoints(amount int) bool // false if insufficient funds
	GrantWeapon(itemID string)
	GrantVehicleAccess(itemID string) // TODO(vehicles owner): wire to vehicles/
}

// Purchase validates and applies a store purchase. Proximity-to-store
// checking (is the player actually standing in a store zone) is expected to
// be done by the caller using world/ before invoking this.
func Purchase(s Spender, itemID string) protocol.PurchaseResultData {
	item, ok := ItemByID(itemID)
	if !ok {
		return protocol.PurchaseResultData{Success: false, ItemID: itemID, Points: s.PointsBalance(), Reason: "unknown item"}
	}
	if s.PointsBalance() < item.Cost {
		return protocol.PurchaseResultData{Success: false, ItemID: itemID, Points: s.PointsBalance(), Reason: "insufficient points"}
	}
	if !s.DeductPoints(item.Cost) {
		return protocol.PurchaseResultData{Success: false, ItemID: itemID, Points: s.PointsBalance(), Reason: "insufficient points"}
	}

	switch item.Kind {
	case "weapon":
		s.GrantWeapon(item.ID)
	case "vehicle":
		s.GrantVehicleAccess(item.ID)
	case "cosmetic":
		// TODO(player owner): grant cosmetic unlock.
	}

	return protocol.PurchaseResultData{Success: true, ItemID: itemID, Points: s.PointsBalance()}
}
