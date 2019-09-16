module github.com/mdblp/hydrophone

go 1.12

replace github.com/tidepool-org/hydrophone => ./

replace github.com/tidepool-org/go-common => github.com/mdblp/go-common v0.1.1-0.20190828100507-09c32bff2777

require (
	github.com/aws/aws-sdk-go v1.19.43
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/gorilla/mux v1.7.2
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af
	github.com/nicksnyder/go-i18n v0.0.0-20181124044605-9c0db6e2b3a5
	github.com/nicksnyder/go-i18n/v2 v2.0.2
	github.com/tidepool-org/go-common v0.0.0-00010101000000-000000000000
	github.com/tidepool-org/hydrophone v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.3.2
	gopkg.in/yaml.v2 v2.2.2
)
