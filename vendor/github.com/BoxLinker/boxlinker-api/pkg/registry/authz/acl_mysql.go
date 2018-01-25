package authz

import (
	"github.com/BoxLinker/boxlinker-api/controller/manager"
	"sync"
	"time"
	"github.com/Sirupsen/logrus"
	"fmt"
	"io"
)

type MysqlACL []MysqlACLEntry
type MysqlACLEntry struct {
	ACLEntry
}

type ACLMysqlConfig struct {
	Manager manager.RegistryManager
	CacheTTL time.Duration
}

type ACLMysqlAuthorizer struct {
	config ACLMysqlConfig
	lastCacheUpdate  time.Time
	staticAuthorizer Authorizer
	manager manager.RegistryManager
	lock sync.RWMutex
	updateTicker     *time.Ticker

}

func NewACLMysqlAuthorizer(config ACLMysqlConfig) (Authorizer, error) {

	authorizer := &ACLMysqlAuthorizer{
		manager: config.Manager,
		updateTicker: time.NewTicker(config.CacheTTL),
		config: config,
	}

	// Initially fetch the ACL from Mysql
	if err := authorizer.updateACLCache(); err != nil {
		return nil, err
	}

	go authorizer.continuouslyUpdateACLCache()


	return authorizer, nil
}

// continuouslyUpdateACLCache checks if the ACL cache has expired and depending
// on the the result it updates the cache with the ACL from the MongoDB server.
// The ACL will be stored inside the static authorizer instance which we use
// to minimize duplication of code and maximize reuse of existing code.
func (ma *ACLMysqlAuthorizer) continuouslyUpdateACLCache() {
	var tick time.Time
	for ; true; tick = <-ma.updateTicker.C {
		aclAge := time.Now().Sub(ma.lastCacheUpdate)
		logrus.Infof("Updating ACL at %s (ACL age: %s. CacheTTL: %s)", tick, aclAge, ma.config.CacheTTL)

		for true {
			err := ma.updateACLCache()
			if err == nil {
				break
			} else if err == io.EOF {
				logrus.Warningf("EOF error received from Mongo. Retrying connection")
				time.Sleep(time.Second)
				continue
			} else {
				logrus.Errorf("Failed to update ACL. ERROR: %s", err)
				logrus.Warningf("Using stale ACL (Age: %s, TTL: %s)", aclAge, ma.config.CacheTTL)
				break
			}
		}
	}
}

func (ma *ACLMysqlAuthorizer) updateACLCache() error {
	var newACL MysqlACL
	acls, err := ma.manager.QueryAllACL()
	if err != nil {
		return err
	}
	for _, a := range acls {
		cond := MatchConditions{
			Account: &a.Account,
			Name: &a.Name,
			Type: &a.Type,
			Service: &a.Service,
			IP: &a.IP,
		}
		if a.Account == "" {
			cond.Account = nil
		}
		if a.Name == "" {
			cond.Name = nil
		}
		if a.Type == "" {
			cond.Type = nil
		}
		if a.Service == "" {
			cond.Service = nil
		}
		if a.IP == "" {
			cond.IP = nil
		}
		newACL = append(newACL, MysqlACLEntry{
			ACLEntry: ACLEntry{
				Actions: &a.ActionsArray,
				Comment: &a.Comment,
				Match: &cond,
			},
		})
	}
	var retACL ACL
	for _, e := range newACL {
		retACL = append(retACL, e.ACLEntry)
	}
	logrus.Debugf("Get all acl: %v", retACL)
	newStaticAuthorizer, err := NewACLAuthorizer(retACL)
	if err != nil {
		return err
	}

	ma.lock.Lock()
	ma.lastCacheUpdate = time.Now()
	ma.staticAuthorizer = newStaticAuthorizer
	ma.lock.Unlock()

	logrus.Infof("Got new ACL from Mysql: %s", retACL)
	logrus.Infof("Installed new ACL from Mysql (%d entries)", len(retACL))

	return nil
}

func (ma *ACLMysqlAuthorizer) Authorize(ai *AuthRequestInfo) ([]string, error) {
	ma.lock.RLock()
	defer ma.lock.RUnlock()

	// Test if authorizer has been initialized
	if ma.staticAuthorizer == nil {
		return nil, fmt.Errorf("MongoDB authorizer is not ready")
	}

	return ma.staticAuthorizer.Authorize(ai)
}

func (acl *ACLMysqlAuthorizer) Stop(){}
func (acl *ACLMysqlAuthorizer) Name() string {
	return "ACLMysqlAuthorizer"
}