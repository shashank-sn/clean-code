package billing

import "context"

type Invoice struct{ ID, TenantID string; Total int }
var invoices = map[string]Invoice{}
func GetInvoice(ctx context.Context, tenantID, id string) (Invoice, error) {
	inv, ok := invoices[id]
	if !ok { return Invoice{}, ErrNotFound }
	return inv, nil
}
var ErrNotFound = errorString("not found")
type errorString string
func (e errorString) Error() string { return string(e) }
