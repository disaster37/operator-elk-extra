package elasticsearchhandler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// LicenseEnableBasic permit to enable basic license
func (h *ElasticsearchHandlerImpl) LicenseEnableBasic() (err error) {
	res, err := h.client.API.License.GetBasicStatus(
		h.client.API.License.GetBasicStatus.WithContext(context.Background()),
		h.client.API.License.GetBasicStatus.WithPretty(),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return errors.Errorf("Error when check if basic license can be enabled: %s", res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	h.log.Debugf("Result when get basic license status: %s", string(b))

	data := make(map[string]interface{})
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	if data["eligible_to_start_basic"] != nil && data["eligible_to_start_basic"].(bool) == false {
		h.log.Infof("Basic license is already enabled")
		return nil
	}
	res, err = h.client.API.License.PostStartBasic(
		h.client.API.License.PostStartBasic.WithContext(context.Background()),
		h.client.API.License.PostStartBasic.WithPretty(),
		h.client.API.License.PostStartBasic.WithAcknowledge(true),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when enable basic license: %s", res.String())
	}

	return nil
}

// LicenseUpdate permit to add or update new license
func (h *ElasticsearchHandlerImpl) LicenseUpdate(license string) (err error) {

	h.log.Debugf("License: %s", license)

	res, err := h.client.API.License.Post(
		h.client.API.License.Post.WithContext(context.Background()),
		h.client.API.License.Post.WithPretty(),
		h.client.API.License.Post.WithAcknowledge(true),
		h.client.API.License.Post.WithBody(strings.NewReader(license)),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add license: %s", res.String())
	}

	return nil
}

// LicenseDelete permit to delete the current license
func (h *ElasticsearchHandlerImpl) LicenseDelete() (err error) {
	res, err := h.client.API.License.Delete(
		h.client.API.License.Delete.WithContext(context.Background()),
		h.client.API.License.Delete.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			h.log.Warnf("License not found, skip it")
			return nil
		}
		return errors.Errorf("Error when delete license: %s", res.String())

	}

	return nil
}

// LicenseGet permit to get the current license
func (h *ElasticsearchHandlerImpl) LicenseGet() (license *olivere.XPackInfoLicense, err error) {

	res, err := h.client.API.License.Get(
		h.client.API.License.Get.WithContext(context.Background()),
		h.client.API.License.Get.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			h.log.Warnf("License not found")
			return nil, nil
		}
		return nil, errors.Errorf("Error when get license: %s", res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	h.log.Debugf("Get license successfully:\n%s", string(b))

	licenseResponse := &olivere.XPackInfoServiceResponse{}
	err = json.Unmarshal(b, licenseResponse)
	if err != nil {
		return nil, err
	}

	return &licenseResponse.License, nil
}

// LicenseDiff permit to compare actual license with expected license.
// It only compare the UID if expected is not basic license
func (h *ElasticsearchHandlerImpl) LicenseDiff(actual, expected *olivere.XPackInfoLicense) (isDiff bool) {

	if actual == nil {
		return true
	}

	// Don't check UID is basic license
	if expected.Type == "basic" {
		return actual.Type != expected.Type
	}

	return actual.UID != expected.UID
}
