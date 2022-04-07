package elasticsearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// UserCreate permit to create new user
func (h *ElasticsearchHandlerImpl) UserCreate(name string, user *olivere.XPackSecurityPutUserRequest) (err error) {

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	res, err := h.client.API.Security.PutUser(
		name,
		bytes.NewReader(data),
		h.client.API.Security.PutUser.WithContext(context.Background()),
		h.client.API.Security.PutUser.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add user %s: %s", name, res.String())
	}

	return nil
}

// UserUpdate permit to update the user
func (h *ElasticsearchHandlerImpl) UserUpdate(name string, user *olivere.XPackSecurityPutUserRequest) (err error) {

	//check if need to update password
	if user.Password != "" || user.PasswordHash != "" {

		payload := make(map[string]string)
		if user.Password != "" {
			payload["password"] = user.Password
		} else {
			payload["password_hash"] = user.PasswordHash
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		res, err := h.client.API.Security.ChangePassword(
			bytes.NewReader(data),
			h.client.API.Security.ChangePassword.WithUsername(name),
			h.client.API.Security.ChangePassword.WithContext(context.Background()),
			h.client.API.Security.ChangePassword.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when change password for user %s: %s", name, res.String())
		}

		h.log.Infof("Updated user password %s successfully", name)
	}

	user.Password = ""
	user.PasswordHash = ""
	return h.UserCreate(name, user)
}

// UserDelete permit to delete the user
func (h *ElasticsearchHandlerImpl) UserDelete(name string) (err error) {

	res, err := h.client.API.Security.DeleteUser(
		name,
		h.client.API.Security.DeleteUser.WithContext(context.Background()),
		h.client.API.Security.DeleteUser.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil

		}
		return errors.Errorf("Error when delete user %s: %s", name, res.String())
	}

	h.log.Infof("Deleted user %s successfully", name)

	return nil
}

// UserGet permot to get the user
func (h *ElasticsearchHandlerImpl) UserGet(name string) (user *olivere.XPackSecurityUser, err error) {

	res, err := h.client.API.Security.GetUser(
		h.client.API.Security.GetUser.WithContext(context.Background()),
		h.client.API.Security.GetUser.WithPretty(),
		h.client.API.Security.GetUser.WithUsername(name),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, nil
		}
		return nil, errors.Errorf("Error when get user %s: %s", name, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	h.log.Debugf("Get user %s successfully:\n%s", name, string(b))
	userResp := make(olivere.XPackSecurityGetUserResponse)
	err = json.Unmarshal(b, &userResp)
	if err != nil {
		return nil, err
	}

	h.log.Infof("Read user %s successfully", name)

	tmp := userResp[name]
	return &tmp, nil
}

// UserDiff permit to check if 2 users are the same
func (h *ElasticsearchHandlerImpl) UserDiff(actual, expected *olivere.XPackSecurityPutUserRequest) (diff string, err error) {
	return standartDiff(actual, expected, h.log, nil)
}
