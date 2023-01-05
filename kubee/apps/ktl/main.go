package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	corev1 "justinsb.com/kubee/api/core/v1"
	metav1 "justinsb.com/kubee/apimachinery/pkg/apis/meta/v1"
	clientcmdv1 "justinsb.com/kubee/machinery/apis/clientcmd/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "unexpected error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}
	p := filepath.Join(homedir, ".kube", "config")
	b, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("reading %q: %w", p, err)
	}

	m := &clientcmdv1.Config{}
	if err := yaml.Unmarshal(b, m); err != nil {
		return fmt.Errorf("parsing %q: %w", p, err)
	}

	restConfig, err := clientcmdv1.GetRESTConfig(ctx, m)
	if err != nil {
		return fmt.Errorf("getting rest config: %w", err)
	}
	// klog.Infof("restConfig is %+v", restConfig)

	client, err := NewClient(restConfig)
	if err != nil {
		return err
	}

	namespaces, err := client.GetNamespaces(ctx).Do()
	if err != nil {
		return err
	}
	klog.Infof("namespaces: %v", protojson.Format(namespaces))

	gv := schema.GroupVersion{
		Group:   "apps",
		Version: "v1",
	}
	resources, err := client.ServerResourcesForGroupVersion(ctx, gv).Do()
	if err != nil {
		return err
	}
	klog.Infof("resources: %v", protojson.Format(resources))
	groups, err := client.Groups(ctx).Do()
	if err != nil {
		return err
	}
	klog.Infof("groups: %v", protojson.Format(groups))

	now := time.Now()
	for _, ns := range namespaces.GetItems() {
		age := now.Sub(ns.GetMetadata().GetCreationTimestamp().Time())
		age = age.Round(time.Second)
		// TODO: Eliminate need for GetMetadata call
		fmt.Printf("%s\t%s\t%v\n", ns.GetMetadata().GetName(), ns.GetStatus().GetPhase(), age)
	}
	return nil
}

type Client struct {
	httpClient *http.Client
	baseURL    url.URL
}

func NewClient(restConfig *clientcmdv1.RESTConfig) (*Client, error) {
	serverURL, err := url.Parse(restConfig.Server)
	if err != nil {
		return nil, fmt.Errorf("error parsing server url %q: %w", restConfig.Server, err)
	}
	tlsConfig := &tls.Config{}

	if restConfig.CertificateAuthorityData != nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(restConfig.CertificateAuthorityData) {
			return nil, fmt.Errorf("error parsing certificate authority data")
		}
		tlsConfig.RootCAs = certPool
	}

	if restConfig.ClientKeyData != nil {
		tlsCert, err := tls.X509KeyPair(restConfig.ClientCertificateData, restConfig.ClientKeyData)
		if err != nil {
			return nil, fmt.Errorf("error parsing client certificate: %w", err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, tlsCert)
	}

	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{
		Transport: httpTransport,
	}
	c := &Client{
		httpClient: httpClient,
		baseURL:    *serverURL,
	}
	return c, nil
}

// Discovery
func (c *Client) ServerResourcesForGroupVersion(ctx context.Context, gv schema.GroupVersion) *Request[metav1.APIResourceList] {
	return buildRequest[metav1.APIResourceList](ctx, c, "apis", gv.Group, gv.Version)
}

func (c *Client) Groups(ctx context.Context) *Request[metav1.APIGroupList] {
	return buildRequest[metav1.APIGroupList](ctx, c, "apis")
}

func (c *Client) GetNamespaces(ctx context.Context) *Request[corev1.NamespaceList] {
	return buildRequest[corev1.NamespaceList](ctx, c, "api", "v1", "namespaces")
}

type Request[T any] struct {
	httpRequest *http.Request
	client      *Client
	err         error
}

func (r *Request[T]) Do() (*T, error) {
	t := new(T)
	err := r.Into(t)
	return t, err
}

func (r *Request[T]) Into(dest *T) error {
	httpResponse, err := r.client.httpClient.Do(r.httpRequest)
	if err != nil {
		return fmt.Errorf("error from http request: %w", err)
	}
	defer httpResponse.Body.Close()

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	klog.V(8).Infof("response is %v", string(body))

	if httpResponse.StatusCode != 200 {
		return fmt.Errorf("unexpected response status %q", httpResponse.Status)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	return nil
}

func buildRequest[T any](ctx context.Context, client *Client, relativePath ...string) *Request[T] {
	endpoint := client.baseURL.JoinPath(relativePath...)
	httpRequest, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	return &Request[T]{
		httpRequest: httpRequest,
		client:      client,
		err:         err,
	}
}
