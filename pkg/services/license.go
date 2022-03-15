package services

import (
	"github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/sirupsen/logrus"
)

type LicenseService interface {
	Reconcile(license *v1alpha1.License) (isCreate, isUpdate bool, err error)
	Delete(license *v1alpha1.License) (err error)
	SetLogger(log *logrus.Entry)
}
