package sesEmailHandler_test

import (
	"testing"

	"github.com/philomusica/tickets-lambda-utils/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-utils/lib/emailHandler/sesEmailHandler"
	"github.com/philomusica/tickets-lambda-utils/lib/paymentHandler"
	"github.com/aws/aws-sdk-go/service/sesv2/sesv2iface"
)

type mockGenerateTicketPDFSuccess struct {
	sesv2iface.SESV2API
}

func TestGenerateTicketPDF(t *testing.T) {
	concert := databaseHandler.Concert{
		ID:               "1234",
		Title:            "Summer Concert",
		ImageURL:         "https://example.com/image.png",
		Location:         "Holy Trinity Church, Longlevens, GL2 0AJ",
		Date:             "Sat 1 Feb, 2023",
		Time:             "7:30pm",
		AvailableTickets: 200,
		FullPrice:        12.0,
		ConcessionPrice:  10.0,
	}

	order := paymentHandler.Order{
		ConcertID:        "1234",
		OrderReference:        "ABC1234",
		FirstName:        "John",
		LastName:         "Smith",
		Email:            "johnsmith@mail.com",
		NumOfFullPrice:   2,
		NumOfConcessions: 1,
	}

	svc := mockGenerateTicketPDFSuccess{}
	sesEmailHandler := sesEmailHandler.New(svc, "tickets@philomusica.org.uk")
	attachment := sesEmailHandler.GenerateTicketPDF(order, concert, true, "https://api.philomusica.org.uk/ticket-redeem")

	if len(attachment) == 0 {
		t.Error("Expected attachment file , got an empty slice")
	}

}
