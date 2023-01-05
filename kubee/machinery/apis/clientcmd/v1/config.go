package v1

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/encoding/prototext"
	"k8s.io/klog/v2"
)

type RESTConfig struct {
	Server string

	CertificateAuthorityData []byte

	ClientCertificateData []byte
	ClientKeyData         []byte
}

func GetRESTConfig(ctx context.Context, config *Config) (*RESTConfig, error) {
	klog.Infof("config is %v", prototext.Format(config))

	currentContextName := config.GetCurrentContext()
	if currentContextName == "" {
		return nil, fmt.Errorf("current-context is not set in kubeconfig")
	}

	var currentContext *Context
	for _, namedContext := range config.Contexts {
		if namedContext.GetName() == currentContextName {
			currentContext = namedContext.GetContext()
		}
	}

	if currentContext == nil {
		return nil, fmt.Errorf("current-context %q was not found", currentContextName)
	}

	var cluster *Cluster
	for _, c := range config.GetClusters() {
		if c.GetName() == currentContext.GetCluster() {
			cluster = c.GetCluster()
		}
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster %q was not found", currentContext.GetCluster())
	}

	var user *AuthInfo
	for _, u := range config.GetUsers() {
		if u.GetName() == currentContext.GetUser() {
			user = u.GetUser()
		}
	}
	if user == nil {
		return nil, fmt.Errorf("user %q was not found", currentContext.GetUser())
	}

	rc := &RESTConfig{
		Server: cluster.GetServer(),
	}

	rc.CertificateAuthorityData = cluster.CertificateAuthorityData
	rc.ClientKeyData = user.ClientKeyData
	rc.ClientCertificateData = user.ClientCertificateData

	return rc, nil
}
