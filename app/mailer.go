package app

import (
	"bytes"
	"errors"
	"text/template"

	"DreamsMoney/feelgoodfoodsnv.com/ordering/persisters"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/repositories"
	"DreamsMoney/feelgoodfoodsnv.com/ordering/templates"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendOrderReceiptEmail(order repositories.Order, orderID persisters.ID) error {
	if order.Customer.Email == "" {
		return errors.New("Order is missing to email")
	}

	htmlTemplate := templates.ParseFiles("./emails/receipt.html", "./pages/footer.html", "./pages/header.html")
	textTemplate := templates.ParseFiles("./emails/receipt.txt")

	toEmail := make(map[string]interface{})
	toEmail["Order"] = order
	toEmail["OrderTotal"] = order.Total()
	toEmail["OrderID"] = orderID

	logo := mail.NewAttachment()
	logo.Content = "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"
	logo.ContentID = "1"
	logo.Filename = "homepage-logo.png"
	logo.Type = "image/png"

	return sendEmail(order.Customer.Email, order.Customer.FirstName, "FeelgoodfoodsNV Order Receipt", "receipt", *textTemplate, *htmlTemplate, toEmail, logo)
}

func sendEmail(toEmail, toName, subject, templateName string, textContent, htmlContent template.Template, data interface{}, attachment *mail.Attachment) error {
	var fromEmail, fromName, mailerKey string
	if serverConfiguration.Environment == "dev" {
		fromEmail = serverConfiguration.MailerFromEmail.Development
		fromName = serverConfiguration.MailerFromName.Development
		mailerKey = serverConfiguration.MailerKeys.Development
	} else {
		fromEmail = serverConfiguration.MailerFromEmail.Production
		fromName = serverConfiguration.MailerFromName.Production
		mailerKey = serverConfiguration.MailerKeys.Production
	}

	emailFrom := mail.NewEmail(fromName, fromEmail)
	emailTo := mail.NewEmail(toName, toEmail)

	var textTemplateBuffer, htmlTemplateBuffer bytes.Buffer
	err := textContent.ExecuteTemplate(&textTemplateBuffer, templateName, data)
	if err != nil {
		return err
	}
	err = htmlContent.ExecuteTemplate(&htmlTemplateBuffer, templateName, data)
	if err != nil {
		return err
	}

	plainTextEmailContent := textTemplateBuffer.String()
	htmlEmailContent := htmlTemplateBuffer.String()

	message := mail.NewSingleEmail(emailFrom, subject, emailTo, plainTextEmailContent, htmlEmailContent)
	if attachment != nil {
		message.AddAttachment(attachment)
	}

	client := sendgrid.NewSendClient(mailerKey)

	response, err := client.Send(message)
	if err != nil {
		return err
	} else if response.StatusCode < 200 || response.StatusCode > 299 {
		return errors.New(response.Body)
	}

	return nil
}
