package emailHandler

import "github.com/philomusica/tickets-lambda-utils/lib/databaseHandler"
import "github.com/philomusica/tickets-lambda-utils/lib/paymentHandler"

type EmailHandler interface {
	GenerateTicketPDF(order paymentHandler.Order, concert databaseHandler.Concert, includeQRCode bool, redeemTicketURL string) (attachment []byte)
	SendEmail(order paymentHandler.Order, attachment []byte) (err error)
	SendPaymentFailureEmail(order paymentHandler.Order) (err error)
}
