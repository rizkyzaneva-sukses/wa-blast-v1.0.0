package handlers

// Semua fitur selalu diizinkan (internal company, tanpa paket langganan).
const planFeatureMessage = ""

func tenantPlanAllows(tenantID uint, feature string) bool {
	return true
}

func agentPlanAllows(agentID uint, feature string) bool {
	return true
}
