package emailHandler

import "aithub.com/philomusica/tickets-lambda-utils/lib/databaseHandler"
import "aithub.com/philomusica/tickets-lambda-utils/lib/paymentHandler"

type EmailHandler interface {
	GenerateTicketPDF(order paymentHandler.Order, concert databaseHandler.Concert, includeQRCode bool, redeemTicketURL string) (attachment []byte)
	SendEmail(order paymentHandler.Order, attachment []byte) (err error)
	SendPaymentFailureEmail(order paymentHandler.Order) (err error)
}
