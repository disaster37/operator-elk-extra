package elasticsearchhandler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	olivere "github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
)

var urlLicense = fmt.Sprintf("%s/_license", baseURL)

func (t *ElasticsearchHandlerTestSuite) TestLicenseGet() {

	// Normale use case
	result := &olivere.XPackInfoServiceResponse{
		License: olivere.XPackInfoLicense{
			UID:  "test",
			Type: "basic",
		},
	}

	httpmock.RegisterResponder("GET", urlLicense, func(req *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, result)
		if err != nil {
			panic(err)
		}
		SetHeaders(resp)
		return resp, nil
	})

	license, err := t.esHandler.LicenseGet()
	if err != nil {
		t.Fail(err.Error())
	}
	assert.Equal(t.T(), "test", license.UID)
	assert.Equal(t.T(), "basic", license.Type)

	// When error
	httpmock.RegisterResponder("GET", urlLicense, httpmock.NewErrorResponder(errors.New("fack error")))
	license, err = t.esHandler.LicenseGet()
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestLicenseDelete() {

	// Normale use case
	httpmock.RegisterResponder("DELETE", urlLicense, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.LicenseDelete()
	if err != nil {
		t.Fail(err.Error())
	}
	// When error
	httpmock.RegisterResponder("DELETE", urlLicense, httpmock.NewErrorResponder(errors.New("Fake error")))
	err = t.esHandler.LicenseDelete()
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestLicenseUpdate() {

	// Normale use case
	httpmock.RegisterResponder("PUT", urlLicense, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.LicenseUpdate("fake license")
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	httpmock.RegisterResponder("PUT", urlLicense, httpmock.NewErrorResponder(errors.New("Fake error")))
	err = t.esHandler.LicenseUpdate("fake license")
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestLicenseEnableBasic() {

	// Normale use case
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/basic_status", urlLicense), func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, `{"eligible_to_start_basic": true}`)
		SetHeaders(resp)
		return resp, nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/start_basic", urlLicense), func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.LicenseEnableBasic()
	if err != nil {
		t.Fail(err.Error())
	}

	// Not eligible
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/basic_status", urlLicense), func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, `{"eligible_to_start_basic": false}`)
		SetHeaders(resp)
		return resp, nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/start_basic", urlLicense), func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})
	err = t.esHandler.LicenseEnableBasic()
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/basic_status", urlLicense), func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, `{"eligible_to_start_basic": true}`)
		SetHeaders(resp)
		return resp, nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/start_basic", urlLicense), httpmock.NewErrorResponder(errors.New("fake error")))
	err = t.esHandler.LicenseEnableBasic()
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestLicenseDiff() {

	// No diff, same UID and not basic
	actual := &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "gold",
	}
	new := &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "gold",
	}

	assert.False(t.T(), t.esHandler.LicenseDiff(actual, new))

	// No diff, basic license
	actual = &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "basic",
	}
	new = &olivere.XPackInfoLicense{
		UID:  "test2",
		Type: "basic",
	}
	assert.False(t.T(), t.esHandler.LicenseDiff(actual, new))

	// Diff, not same id and not basic
	actual = &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "gold",
	}
	new = &olivere.XPackInfoLicense{
		UID:  "test2",
		Type: "gold",
	}
	assert.True(t.T(), t.esHandler.LicenseDiff(actual, new))

	// Diff, not same license type
	actual = &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "gold",
	}
	new = &olivere.XPackInfoLicense{
		UID:  "test2",
		Type: "basic",
	}
	assert.True(t.T(), t.esHandler.LicenseDiff(actual, new))

}
