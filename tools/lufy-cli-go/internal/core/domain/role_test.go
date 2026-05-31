package domain

import "testing"

func TestDefaultRoleContractsCoverAllRoles(t *testing.T) {
	contracts := DefaultRoleContracts()
	got := map[RoleID]bool{}
	for _, contract := range contracts {
		got[contract.ID] = true
		if contract.Output.Schema != "result-contract/v1" {
			t.Fatalf("%s schema = %s", contract.ID, contract.Output.Schema)
		}
		if len(contract.Output.AllowedStatus) == 0 {
			t.Fatalf("%s missing allowed status", contract.ID)
		}
	}

	for _, role := range []RoleID{RoleOrchestrator, RoleRouter, RoleExplorer, RoleImplementer, RoleTestWriter, RoleValidator, RoleReviewer, RoleDelivery} {
		if !got[role] {
			t.Fatalf("missing role contract for %s", role)
		}
	}
}

func TestDeliveryContractCarriesPRContentSkillSlot(t *testing.T) {
	for _, contract := range DefaultRoleContracts() {
		if contract.ID != RoleDelivery {
			continue
		}
		for _, slot := range contract.DirectSlots {
			if slot == SkillSlotDeliveryPRContent {
				return
			}
		}
		t.Fatalf("delivery contract must include %s", SkillSlotDeliveryPRContent)
	}
	t.Fatalf("delivery contract not found")
}
