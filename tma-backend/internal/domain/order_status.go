package domain

var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusNew:                 {OrderStatusWaitingPayment, OrderStatusCancelled},
	OrderStatusWaitingPayment:      {OrderStatusPaymentVerification, OrderStatusCancelled},
	OrderStatusPaymentVerification: {OrderStatusPaid, OrderStatusCancelled, OrderStatusRefundRequested},
	OrderStatusPaid:                {OrderStatusKeyIssued, OrderStatusWaitingActivation, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusWaitingActivation:   {OrderStatusAwaitingCredentials, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusAwaitingCredentials: {OrderStatusCredentialsReceived, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusCredentialsReceived: {OrderStatusAwaiting2FA, OrderStatusAwaitingCredentials, OrderStatusCredentialsInvalid, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusCredentialsInvalid:  {OrderStatusAwaitingCredentials, OrderStatusCredentialsReceived, OrderStatusCancelled},
	OrderStatusAwaiting2FA:         {OrderStatusActivating, OrderStatusInvalid2FA, OrderStatusCredentialsReceived, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusInvalid2FA:          {OrderStatusAwaiting2FA, OrderStatusActivating, OrderStatusCancelled},
	OrderStatusActivating:          {OrderStatusActivated, OrderStatusInvalid2FA, OrderStatusRefundRequested, OrderStatusCancelled},
	OrderStatusActivated:           {OrderStatusCompleted, OrderStatusCancelled},
	OrderStatusKeyIssued:           {OrderStatusCompleted, OrderStatusCancelled},
	OrderStatusRefundRequested:     {OrderStatusRefunded, OrderStatusCancelled},
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
