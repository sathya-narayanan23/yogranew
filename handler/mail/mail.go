package mail

import(
	"fmt"
	"net/smtp"
	"time"
	"gopkg.in/gomail.v2"
)


func SendPlanExpirationEmail(email, plan string, planExpiration time.Time) error {
	from := "rts.sathyanarayanan@gmail.com"
	password := "pvnkounjkpghyyxb"
	to := email

	// Format the plan expiration date
	planExpirationFormatted := planExpiration.Format("2006-01-02")

	msg := fmt.Sprintf("Currently your plan is upgraded to  %s,\n Your plan validity through %s .", plan, planExpirationFormatted)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func SendEmailNotification(role, name, email, mobileNumber, passwords string) error {
	from := "rts.sathyanarayanan@gmail.com"
	password := "pvnkounjkpghyyxb"
	to := email
	msg := fmt.Sprintf("\n\n Resource  has been created successfully.\n  %s %s  login Credentials.\n mobileNumber:%s .\n password:%s", role, name, mobileNumber, passwords)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func SendEmailNotificationss(email, plan string, planExpiration time.Time) error {
	from := "rts.sathyanarayanan@gmail.com"
	password := "pvnkounjkpghyyxb"
	to := email

	// Format the plan expiration date
	planExpirationFormatted := planExpiration.Format("2006-01-02")

	msg := fmt.Sprintf("Your Account has been created successfully in the Yogra app.\n\n"+
		"Your Current Plan is: %s\n"+
		"Plan Expiration Date: %s\n", plan, planExpirationFormatted)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}


func SendEmailNotificationnew(email, plan string, planExpiration time.Time) error {
	from := "rts.sathyanarayanan@gmail.com"
	password := "pvnkounjkpghyyxb"
	to := email

	// Format the plan expiration date
	planExpirationFormatted := planExpiration.Format("2006-01-02")

	msg := fmt.Sprintf("Your Yogra plan in expiring soon  .\n\n"+
		"Your Current Plan is: %s\n"+
		"Plan Expiration Date: %s\n", plan, planExpirationFormatted)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}


// subject := "YOGRA LICENSE "
// body := "Your account has been created  successfully."

func sendEmailWithAttachment(email, subject, body, filePath string) error {
	m := gomail.NewMessage()

	// Sender's email address
	m.SetHeader("From", "rts.sathyanarayanan@gmail.com")

	// Recipient's email address
	m.SetHeader("To", email)

	// Email subject
	m.SetHeader("Subject", subject)

	// Attach the file to the email
	m.Attach(filePath)

	// Email body
	m.SetBody("text/plain", body)

	// Create a new message using the specified settings
	d := gomail.NewDialer("smtp.gmail.com", 587, "rts.sathyanarayanan@gmail.com", "pvnkounjkpghyyxb")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
