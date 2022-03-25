package elasticsearchhandler

import (
	elastic "github.com/elastic/go-elasticsearch/v8"
	olivere "github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type ElasticsearchHandler interface {
	// License scope
	LicenseUpdate(license string) (err error)
	LicenseDelete() (err error)
	LicenseGet() (license *olivere.XPackInfoLicense, err error)
	LicenseDiff(actual, expected *olivere.XPackInfoLicense) (diff bool)
	LicenseEnableBasic() (err error)

	// ILM scope
	ILMUpdate(name, rawPolicy string) (err error)
	ILMDelete(name string) (err error)
	ILMGet(name string) (policy map[string]any, err error)
	ILMDiff(actual, expected map[string]any) (diff string, err error)

	// SLM scope
	SLMUpdate(name, rawPolicy string) (err error)
	SLMDelete(name string) (err error)
	SLMGet(name string) (policy *SnapshotLifecyclePolicySpec, err error)
	SLMDiff(actual, expected *SnapshotLifecyclePolicySpec) (diff string, err error)

	// Snapshot repository scope
	SnapshotRepositoryUpdate(name string, repository *olivere.SnapshotRepositoryMetaData) (err error)
	SnapshotRepositoryDelete(name string) (err error)
	SnapshotRepositoryGet(name string) (repository *olivere.SnapshotRepositoryMetaData, err error)
	SnapshotRepositoryDiff(actual, expected *olivere.SnapshotRepositoryMetaData) (diff string, err error)

	// Role scope
	RoleUpdate(name string, role *olivere.XPackSecurityRole) (err error)
	RoleDelete(name string) (err error)
	RoleGet(name string) (role *olivere.XPackSecurityRole, err error)
	RoleDiff(actual, expected *olivere.XPackSecurityRole) (diff string, err error)

	SetLogger(log *logrus.Entry)
}

type ElasticsearchHandlerImpl struct {
	client *elastic.Client
	log    *logrus.Entry
}

func NewElasticsearchHandler(cfg elastic.Config, log *logrus.Entry) (ElasticsearchHandler, error) {

	client, err := elastic.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ElasticsearchHandlerImpl{
		client: client,
		log:    log,
	}, nil
}

func (h *ElasticsearchHandlerImpl) SetLogger(log *logrus.Entry) {
	h.log = log
}
