package entities

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-mongodb/src/connection"
	"github.com/newrelic/nri-mongodb/src/metrics"
	"strings"
)

type MongodCollector struct {
	HostCollector
}

func (c MongodCollector) GetEntity(i *integration.Integration) (*integration.Entity, error) {
	return i.Entity(c.ConnectionInfo.Host, "mongod")
}

func (c MongodCollector) CollectMetrics(e *integration.Entity) {
	session, err := c.ConnectionInfo.CreateSession()
	if err != nil {
		log.Error("Failed to connect to %s: %v", c.ConnectionInfo.Host, err)
		return
	}

	ms := e.NewMetricSet("MongodSample",
		metric.Attribute{Key: "displayName", Value: e.Metadata.Name},
		metric.Attribute{Key: "entityName", Value: fmt.Sprintf("%s:%s", e.Metadata.Namespace, e.Metadata.Name)},
	)

	var isMaster metrics.IsMaster
	err = session.Run(map[interface{}]interface{}{"isMaster": 1}, &isMaster)

	if err := ms.MarshalMetrics(isMaster); err != nil {
		log.Error("Failed to marshal isMaster metrics for entity %s: %v", e.Metadata.Name, err)

	}

	if isMaster.SetName != "" {
		if err := collectReplSetMetrics(ms, c.ConnectionInfo, session); err != nil {
			log.Error("Failed to collect repl set metrics for entity %s: %v", e.Metadata.Name, err)
		}
	}

	var ss metrics.ServerStatus
	if err := session.Run(map[interface{}]interface{}{"serverStatus": 1}, &ss); err != nil {
		log.Error("Failed to collect serverStatus metrics for entity %s: %v", e.Metadata.Name, err)
	}

	if err := ms.MarshalMetrics(ss); err != nil {
		log.Error("Failed to marshal metrics for entity %s: %v", e.Metadata.Name, err)
	}

}

func GetMongods(shard *ShardCollector) ([]*MongodCollector, error) {
	hostPorts, _ := parseReplicaSetString(shard.Host)

	var mongodCollectors []*MongodCollector
	for _, hostPort := range hostPorts {
		ci := connection.DefaultConnectionInfo()
		ci.Host = hostPort.Host
		ci.Port = hostPort.Port

		newMongodCollector := &MongodCollector{
			HostCollector{ConnectionInfo: ci},
		}
		mongodCollectors = append(mongodCollectors, newMongodCollector)
	}

	return mongodCollectors, nil
}

func collectReplSetMetrics(ms *metric.Set, c *connection.ConnectionInfo, session *mgo.Session) error {

	var replSetStatus metrics.ReplSetGetStatus
	err := session.Run(map[interface{}]interface{}{"replSetGetStatus": 1}, &replSetStatus)
	if err != nil {
		return err
	}

	for _, host := range replSetStatus.Members {
		if strings.HasPrefix(*host.Name, c.Host) {

		}
	}

	return nil

}
