package sesEmailHandler

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/aws/aws-sdk-go/service/sesv2/sesv2iface"
	"github.com/philomusica/tickets-lambda-utils/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-utils/lib/emailHandler"
	"github.com/philomusica/tickets-lambda-utils/lib/paymentHandler"
	"github.com/signintech/gopdf"
	"github.com/skip2/go-qrcode"
	"gopkg.in/gomail.v2"
)

// ===============================================================================================================================
// TYPE DEFINITIONS
// ===============================================================================================================================

type SESEmailHandler struct {
	svc           sesv2iface.SESV2API
	senderAddress string
}

// ===============================================================================================================================
// END TYPE DEFINITIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PRIVATE FUNCTIONS
// ===============================================================================================================================

// addDetailToPDF takes a pointer to a Go.Pdf struct and key and value strings, and writes them to the pdf
func addDetailToPDF(pdf *gopdf.GoPdf, key string, value string) {
	yPos := pdf.GetY() + 30.0
	pdf.SetXY(pdf.MarginLeft(), yPos)
	pdf.SetFont("nunito-light", "", 11)
	pdf.Cell(nil, fmt.Sprintf("%s:", key))
	yPos += 15.0
	pdf.SetXY(pdf.MarginLeft(), yPos)
	pdf.SetFont("nunito", "", 14)
	pdf.Cell(nil, value)
}

// buildAdmitString takes an order struct and returns a formatted string indicating how many people to admit (e.g. "2 adults and 1 concession")
func buildAdmitString(order paymentHandler.Order) string {
	var admitString strings.Builder
	if order.NumOfFullPrice > 0 {
		admitString.WriteString(fmt.Sprintf("%d ", order.NumOfFullPrice))
		var ticketType string
		if order.NumOfFullPrice == 1 {
			ticketType = "adult"
		} else {
			ticketType = "adults"
		}
		admitString.WriteString(ticketType)
	}

	if order.NumOfConcessions > 0 {
		if admitString.Len() > 0 {
			admitString.WriteString(" and ")
		}
		admitString.WriteString(fmt.Sprintf("%d ", order.NumOfConcessions))
		var ticketType string
		if order.NumOfConcessions == 1 {
			ticketType = "concession"
		} else {
			ticketType = "concessions"
		}
		admitString.WriteString(ticketType)
	}
	return admitString.String()
}

// ===============================================================================================================================
// END PRIVATE FUNCTIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PUBLIC FUNCTIONS
// ===============================================================================================================================

func (s SESEmailHandler) CreateCalendarInvites(title string, location string, start int64, description string) (calLinks emailHandler.CalendarLinks) {
	escTitle := url.QueryEscape(title)
	escLocation := url.QueryEscape(location)
	escDescription := url.QueryEscape(description)

	tStart := time.Unix(start, 0)
	tEnd := tStart.Add(time.Hour * 2)
	
	gCalFormat := "20060102T150405Z"
	gDates := url.QueryEscape(fmt.Sprintf("%s/%s", tStart.Format(gCalFormat), tEnd.Format(gCalFormat)))

	outEnd := url.QueryEscape(tEnd.Format(time.RFC3339))
	outStart := url.QueryEscape(tStart.Format(time.RFC3339))

	calLinks.GoogleCalendarLink = fmt.Sprintf("https://calendar.google.com/calendar/render?action=TEMPLATE&dates=%s&details=%s&location=%s&text=%s", gDates, escDescription, escLocation, escTitle)
	calLinks.OutlookCalendarLink = fmt.Sprintf("https://outlook.live.com/calendar/0/action/compose?body=%s&enddt=%s&location=%s&path=%%2Fcalendar%%2Faction%%2Fcompose&rru=addevent&startdt=%s&subject=%s", escDescription, outEnd, escLocation, outStart, escTitle)

	// Construct the ICS file contents
	var sb bytes.Buffer
	sb.WriteString("BEGIN:VCALENDAR\n")
	sb.WriteString("VERSION:2.0\n")
	sb.WriteString("BEGIN:VEVENT\n")
	sb.WriteString(fmt.Sprintf("DTSTART:%s\n", tStart.Format(gCalFormat)))
	sb.WriteString(fmt.Sprintf("DTEND:%s\n", tEnd.Format(gCalFormat)))
	sb.WriteString(fmt.Sprintf("SUMMARY:%s\n", title))
	sb.WriteString(fmt.Sprintf("DESCRIPTION:%s\n", description))
	sb.WriteString(fmt.Sprintf("LOCATION:%s\n", location))
	sb.WriteString("END:VEVENT\n")
	sb.WriteString("END:VCALENDAR\n")

	calLinks.ICSFile = sb.Bytes()

	return
}
// GenerateTicketPDF takes an order struct and returns a PDF file in a byte array and an error if fails
func (s SESEmailHandler) GenerateTicketPDF(order paymentHandler.Order, concert databaseHandler.Concert, includeQRCode bool) (attachment []byte) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	marginSize := 30.0
	pdf.SetMargins(marginSize, marginSize, marginSize, marginSize)

	// Load fonts font
	pdf.AddTTFFont("nunito", "./Nunito-Regular.ttf")
	pdf.AddTTFFont("nunito-bold", "./Nunito-Bold.ttf")
	pdf.AddTTFFont("nunito-light", "./Nunito-Light.ttf")

	// Add header
	pdf.AddHeader(func() {
		pdf.SetFont("nunito", "", 14)
		pdf.Cell(nil, "PHILOMUSICA PRESENTS:")
		pdf.SetXY(marginSize, marginSize+25.0)
		pdf.SetFont("nunito-bold", "", 20)
		pdf.Cell(nil, strings.ToUpper(concert.Title))
	})

	// Add one page
	pdf.AddPage()

	pdf.SetY(pdf.GetY() + 30.0)
	addDetailToPDF(&pdf, "Location", concert.Location)

	if includeQRCode {
		var qrcodeImage []byte
		qrcodeImage, _ = qrcode.Encode(fmt.Sprintf("%s-%s", concert.ID, order.OrderReference), qrcode.Medium, 360)
		ih, _ := gopdf.ImageHolderByBytes(qrcodeImage)
		pdf.ImageByHolder(ih, gopdf.PageSizeA4.W/2+30.0, pdf.GetY(), nil)
	}

	addDetailToPDF(&pdf, "Date", fmt.Sprintf("%s @ %s", concert.Date, concert.Time))
	addDetailToPDF(&pdf, "Name", fmt.Sprintf("%s %s", order.FirstName, order.LastName))
	addDetailToPDF(&pdf, "Reference", order.OrderReference)
	addDetailToPDF(&pdf, "Admit", buildAdmitString(order))
	return pdf.GetBytesPdf()
}

// New takes an SES V2 interface and sender email address and returns a newly created SESEmailHandler struct
func New(svc sesv2iface.SESV2API, senderAddress string) SESEmailHandler {
	return SESEmailHandler{
		svc,
		senderAddress,
	}
}

// SendEmail takes an order struct and attachment (in bytes) and sends an email to the customer, using the AWS SES v2 API. Returns an error if fails, or nil if successful
func (s SESEmailHandler) SendEmail(order paymentHandler.Order, attachment []byte, calendarLinks emailHandler.CalendarLinks) (err error) {
	icsFileName := "concert.ics"
	msg := gomail.NewMessage()
	msg.SetHeader("To", order.Email)
	msg.SetHeader("From", s.senderAddress)
	msg.SetHeader("Subject", "Order Confirmation")
	msg.SetBody("text/html", fmt.Sprintf("<div>Dear %s</div><br><div>Many thanks for purchasing tickets to Philomusica's concert. Your eTicket is attached as a PDF to this email. Please bring this PDF with you to the concert, either in digital or paper form.</div><br><div>We look forward to seeing you there!</div><div>Philomusica</div><br><br><a href=\"%s\">Add to Google calendar</a><br><a href=\"%s\">Add to Outlook calendar</a><br><div>For Apple's iCalendar, please download the %s file and upload it to your calendar</div><div>Please consider the environment before printing your eTicket</div>", order.FirstName, calendarLinks.GoogleCalendarLink, calendarLinks.OutlookCalendarLink, icsFileName))
	msg.Attach(
		"philomusica-concert-tickets.pdf",
		gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(attachment)
			return err
		}),
		gomail.SetHeader(map[string][]string{"Content-Type": {"application/pdf"}}),
	)

	msg.Attach(
		icsFileName,
		gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(calendarLinks.ICSFile)
			return err
		}),
		gomail.SetHeader(map[string][]string{"Content-Type": {"text/calendar"}}),
	)


	var rawEmail bytes.Buffer
	msg.WriteTo(&rawEmail)

	// send raw email
	_, err = s.svc.SendEmail(
		&sesv2.SendEmailInput{
			Content: &sesv2.EmailContent{
				Raw: &sesv2.RawMessage{Data: rawEmail.Bytes()},
			},
			Destination: &sesv2.Destination{
				ToAddresses: []*string{&order.Email},
			},
			FromEmailAddress: &s.senderAddress,
		},
	)

	return
}

func (s SESEmailHandler) SendPaymentFailureEmail(order paymentHandler.Order) (err error) {
	subject := "Ticket payment failed"
	htmlBody := fmt.Sprintf(`<div>Dear %s</div><br><div>Unfortuantely we were unable to process your payment. Please try again, and if you continue to have problems please <a href="https://philomusica.org.uk/contact.html">contact us</a></div><br><div>Many thanks</div><div>Philomusica</div>`, order.FirstName)

	_, err = s.svc.SendEmail(
		&sesv2.SendEmailInput{
			Content: &sesv2.EmailContent{
				Simple: &sesv2.Message{
					Body: &sesv2.Body{
						Html: &sesv2.Content{
							Data: aws.String(htmlBody),
						},
					},
					Subject: &sesv2.Content{
						Data: aws.String(subject),
					},
				},
			},
			Destination: &sesv2.Destination{
				ToAddresses: []*string{&order.Email},
			},
			FromEmailAddress: &s.senderAddress,
		},
	)
	return
}

// ===============================================================================================================================
// END PUBLIC FUNCTIONS
// ===============================================================================================================================
