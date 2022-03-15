package elasticsearchhandler

import (
	"net/http"

	olivere "github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
)

func (t *ElasticsearchHandlerTestSuite) TestLicenseGet() {

	mock := &MockTransport{}
	esHandler := GetMockElasticsearchHandler(mock)

	// Normale use case
	result := &olivere.XPackInfoServiceResponse{
		License: olivere.XPackInfoLicense{
			UID:  "test",
			Type: "basic",
		},
	}
	mock.SetResponse(http.StatusOK, result, false)

	license, err := esHandler.LicenseGet()
	if err != nil {
		t.Fail(err.Error())
	}
	assert.Equal(t.T(), "test", license.UID)
	assert.Equal(t.T(), "basic", license.Type)

	// When error
	mock.SetResponse(http.StatusOK, result, true)
	license, err = esHandler.LicenseGet()
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestLicenseDelete() {

	mock := &MockTransport{}
	esHandler := GetMockElasticsearchHandler(mock)

	// Normale use case
	mock.SetResponse(http.StatusOK, map[string]string{}, false)

	err := esHandler.LicenseDelete()
	if err != nil {
		t.Fail(err.Error())
	}
	// When error
	mock.SetResponse(http.StatusOK, map[string]string{}, true)
	err = esHandler.LicenseDelete()
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestLicenseUpdate() {

	mock := &MockTransport{}
	esHandler := GetMockElasticsearchHandler(mock)

	// Normale use case
	mock.SetResponse(http.StatusOK, map[string]string{}, false)

	err := esHandler.LicenseUpdate("fake license")
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	mock.SetResponse(http.StatusOK, map[string]string{}, true)
	err = esHandler.LicenseUpdate("fake license")
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestLicenseEnableBasic() {

	mock := &MockTransport{}
	esHandler := GetMockElasticsearchHandler(mock)

	// Normale use case
	mock.SetResponse(http.StatusOK, map[string]interface{}{"eligible_to_start_basic": true}, false)

	err := esHandler.LicenseEnableBasic()
	if err != nil {
		t.Fail(err.Error())
	}

	// Nor eligible
	mock.SetResponse(http.StatusOK, map[string]interface{}{"eligible_to_start_basic": false}, false)

	err = esHandler.LicenseEnableBasic()
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	mock.SetResponse(http.StatusOK, map[string]string{}, true)
	err = esHandler.LicenseEnableBasic()
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestLicenseDiff() {

	mock := &MockTransport{}
	esHandler := GetMockElasticsearchHandler(mock)

	// No diff, same UID and not basic
	actual := &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "gold",
	}
	new := &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "gold",
	}

	assert.False(t.T(), esHandler.LicenseDiff(actual, new))

	// No diff, basic license
	actual = &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "basic",
	}
	new = &olivere.XPackInfoLicense{
		UID:  "test2",
		Type: "basic",
	}
	assert.False(t.T(), esHandler.LicenseDiff(actual, new))

	// Diff, not same id and not basic
	actual = &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "gold",
	}
	new = &olivere.XPackInfoLicense{
		UID:  "test2",
		Type: "gold",
	}
	assert.True(t.T(), esHandler.LicenseDiff(actual, new))

	// Diff, not same license type
	actual = &olivere.XPackInfoLicense{
		UID:  "test",
		Type: "gold",
	}
	new = &olivere.XPackInfoLicense{
		UID:  "test2",
		Type: "basic",
	}
	assert.True(t.T(), esHandler.LicenseDiff(actual, new))

}
