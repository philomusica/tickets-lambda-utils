package emailHandler

import "github.com/philomusica/tickets-lambda-utils/lib/databaseHandler"
import "github.com/philomusica/tickets-lambda-utils/lib/paymentHandler"

type CalendarLinks struct {
	GoogleCalendarLink string
	OutlookCalendarLink string
	ICSFile []byte
}

type EmailHandler interface {
	CreateCalendarInvites(title string, location string, start int64, description string) (calLinks CalendarLinks)
	GenerateTicketPDF(order paymentHandler.Order, concert databaseHandler.Concert, includeQRCode bool) (attachment []byte)
	SendEmail(order paymentHandler.Order, attachment []byte, calendarLinks CalendarLinks) (err error)
	SendPaymentFailureEmail(order paymentHandler.Order) (err error)
}
