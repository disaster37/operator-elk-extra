package elasticsearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// RoleUpdate permit to update role
func (h *ElasticsearchHandlerImpl) RoleUpdate(name string, role *olivere.XPackSecurityRole) (err error) {

	data, err := json.Marshal(role)
	if err != nil {
		return err
	}

	res, err := h.client.API.Security.PutRole(
		name,
		bytes.NewReader(data),
		h.client.API.Security.PutRole.WithContext(context.Background()),
		h.client.API.Security.PutRole.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add role %s: %s\ndata: %s", name, res.String(), string(data))
	}

	return nil
}

// RoleDelete permit to delete role
func (h *ElasticsearchHandlerImpl) RoleDelete(name string) (err error) {

	res, err := h.client.API.Security.DeleteRole(
		name,
		h.client.API.Security.DeleteRole.WithContext(context.Background()),
		h.client.API.Security.DeleteRole.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil

		}
		return errors.Errorf("Error when delete role %s: %s", name, res.String())
	}

	h.log.Infof("Deleted role %s successfully", name)

	return nil
}

// RoleGet permit to get role
func (h *ElasticsearchHandlerImpl) RoleGet(name string) (role *olivere.XPackSecurityRole, err error) {

	res, err := h.client.API.Security.GetRole(
		h.client.API.Security.GetRole.WithContext(context.Background()),
		h.client.API.Security.GetRole.WithPretty(),
		h.client.API.Security.GetRole.WithName(name),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, nil
		}
		return nil, errors.Errorf("Error when get role %s: %s", name, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	h.log.Debugf("Get role %s successfully:\n%s", name, string(b))
	roleResp := make(olivere.XPackSecurityGetRoleResponse)
	err = json.Unmarshal(b, &roleResp)
	if err != nil {
		return nil, err
	}

	tmp := roleResp[name]

	return &tmp, nil
}

// RoleDiff permit to check if 2 role are the same
func (h *ElasticsearchHandlerImpl) RoleDiff(actual, expected *olivere.XPackSecurityRole) (diff string, err error) {
	return standartDiff(&actual, &expected, h.log)
}
