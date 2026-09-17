package billing

import "context"

type Claims struct{ TenantID string }
type Request struct{ Claims Claims; InvoiceID string }

func Handle(ctx context.Context, req Request) (Invoice, error) {
	return GetInvoice(ctx, req.Claims.TenantID, req.InvoiceID)
}
