package elasticsearchhandler

import (
	"encoding/json"

	"github.com/elastic/go-ucfg"
	ucfgjson "github.com/elastic/go-ucfg/json"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func standartDiff(actual, expected any, log *logrus.Entry) (diff string, err error) {
	acualByte, err := json.Marshal(actual)
	if err != nil {
		return diff, err
	}
	expectedByte, err := json.Marshal(expected)
	if err != nil {
		return diff, err
	}

	actualConf, err := ucfgjson.NewConfig(acualByte, ucfg.PathSep("."))
	if err != nil {
		log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), string(acualByte))
		return diff, err
	}
	if err = actualConf.Unpack(actual); err != nil {
		return diff, err
	}
	expectedConf, err := ucfgjson.NewConfig(expectedByte, ucfg.PathSep("."))
	if err != nil {
		log.Errorf("Error when converting new Json: %s\ndata: %s", err.Error(), string(expectedByte))
		return diff, err
	}
	if err = expectedConf.Unpack(expected); err != nil {
		return diff, err
	}

	return cmp.Diff(actual, expected), nil
}
