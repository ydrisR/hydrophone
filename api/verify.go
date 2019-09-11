package api

import (
	"log"
	"net/http"
	"regexp"
)

// Send a test email to prove all configuration is in place for sending emails
//
// status: 200
// status: 500 internal server error
func (a *Api) sendSanityCheckEmail(res http.ResponseWriter, req *http.Request, vars map[string]string) {

	log.Printf("Sanity check email route")

	var recipient string = a.Config.TestEmail
	var subject string = "Sanity Check Email"
	var body string = "This is an automatic email sent from Hydrophone service to prove all configuration is in place for sending emails"

	// To ensure the route is not used for spamming, we ensure requestor has a valid server token
	if token := a.token(res, req); token == nil {
		log.Printf("No valid token is found")
		return
	}

	// Valid valid email address is found
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if recipient == "" || !re.MatchString(recipient) {
		log.Printf("No valid email for sanity check is found")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Try sending
	if status, details := a.notifier.Send([]string{a.Config.TestEmail}, subject, body); status != http.StatusOK {
		log.Printf("Issue sending sanity check email: Status [%d] Message [%s]", status, details)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Success: sanity check email successfully sent to [%s]", recipient)
	//unless no email was given we say its all good
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(STATUS_OK))
	return
}
