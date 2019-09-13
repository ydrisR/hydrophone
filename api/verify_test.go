package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestSanityCheckResponds(t *testing.T) {

	tests := []toTest{
		{
			// no test email in config
			method:    "POST",
			url:       "/sanity_check",
			token:     "any.token.will.do.because.of.how.shoreline.mock.is.built",
			testEmail: "",
			respCode:  500,
		},
		{
			// malformed email in config
			method:    "POST",
			url:       "/sanity_check",
			token:     "any.token.will.do.because.of.how.shoreline.mock.is.built",
			testEmail: "ff@ss",
			respCode:  500,
		},
		{
			// malformed email in config
			method:    "POST",
			url:       "/sanity_check",
			token:     "any.token.will.do.because.of.how.shoreline.mock.is.built",
			testEmail: "ffss.com",
			respCode:  500,
		},
		{
			// return 401 if no session token is present
			method:    "POST",
			url:       "/sanity_check",
			testEmail: "ff@ss.com",
			respCode:  401,
		},
		{
			// return 200 if everything is in place
			method:    "POST",
			url:       "/sanity_check",
			testEmail: "ff@ss.com",
			token:     "any.token.will.do.because.of.how.shoreline.mock.is.built",
			respCode:  200,
		},
	}

	for idx, test := range tests {

		//fresh each time
		var testRtr = mux.NewRouter()

		FAKE_CONFIG.TestEmail = test.testEmail

		hydrophone := InitApi(FAKE_CONFIG, mockStore, mockNotifier, mockShoreline, mockGatekeeper, mockMetrics, mockSeagull, mockTemplates)
		hydrophone.SetHandlers("", testRtr)

		var body = &bytes.Buffer{}
		// build the body only if there is one defined in the test
		if len(test.body) != 0 {
			json.NewEncoder(body).Encode(test.body)
		}
		request, _ := http.NewRequest(test.method, test.url, body)
		if test.token != "" {
			request.Header.Set(TP_SESSION_TOKEN, test.token)
		}
		response := httptest.NewRecorder()
		testRtr.ServeHTTP(response, request)

		if response.Code != test.respCode {
			t.Fatalf("Test %d url: '%s'\nNon-expected status code %d (expected %d):\n\tbody: %v",
				idx, test.url, response.Code, test.respCode, response.Body)
		}

		if response.Body.Len() != 0 && len(test.response) != 0 {
			// compare bodies by comparing the unmarshalled JSON results
			var result = &testJSONObject{}

			if err := json.NewDecoder(response.Body).Decode(result); err != nil {
				t.Logf("Err decoding nonempty response body: [%v]\n [%v]\n", err, response.Body)
				return
			}

			if cmp := result.deepCompare(&test.response); cmp != "" {
				t.Fatalf("Test %d url: '%s'\n\t%s\n", idx, test.url, cmp)
			}
		}
	}
}
