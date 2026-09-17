package payments

func Checkout(key string, cents int) (string, error) { return CreatePayment(key, cents) }
