package domain

var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusNew:                 {OrderStatusWaitingPayment, OrderStatusCancelled},
	OrderStatusWaitingPayment:      {OrderStatusPaymentVerification, OrderStatusCancelled},
	OrderStatusPaymentVerification: {OrderStatusPaid, OrderStatusCancelled, OrderStatusRefundRequested},
	OrderStatusPaid:                {OrderStatusKeyIssued, OrderStatusWaitingActivation, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusWaitingActivation:   {OrderStatusAwaitingCredentials, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusAwaitingCredentials: {OrderStatusCredentialsReceived, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusCredentialsReceived: {OrderStatusAwaiting2FA, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusAwaiting2FA:         {OrderStatusActivating, OrderStatusCredentialsReceived, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusActivating:          {OrderStatusActivated, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusActivated:           {OrderStatusCompleted},
	OrderStatusKeyIssued:           {OrderStatusCompleted},
	OrderStatusRefundRequested:     {OrderStatusRefunded},
}

func IsValidTransition(from, to OrderStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
