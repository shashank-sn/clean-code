package billing

import "testing"

func TestTenantBoundary(t *testing.T) {
	invoices = map[string]Invoice{"inv-1": {ID: "inv-1", TenantID: "tenant-a", Total: 42}}
	if _, err := GetInvoice(nil, "tenant-b", "inv-1"); err == nil { t.Fatal("foreign tenant received an invoice") }
	if got, err := GetInvoice(nil, "tenant-a", "inv-1"); err != nil || got.Total != 42 { t.Fatalf("owner lookup failed: %#v %v", got, err) }
}
