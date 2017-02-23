package apiv1

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/runner-mei/gowbem"
)

type SMIS struct {
	host      string
	port      string
	insecure  bool
	username  string
	password  string
	namespace string
	client    *http.Client
}

func New(host string, port string, insecure bool, username string, password string) (*SMIS, error) {
	if host == "" || port == "" || username == "" || password == "" {
		return nil, errors.New("Missing host (SMIS Host IP), port (SMIS Host Port), username, or password \n Check Environment Variables..")
	}
	namespace := "root/emc"

	var client *http.Client
	if insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	return &SMIS{host, port, insecure, username, password, namespace, client}, nil
}

func getArrayUrl(smis *SMIS) *url.URL {
	var schema string
	if smis.insecure {
		schema = "http"
	} else {
		schema = "https"
	}
	return &url.URL{
		Scheme: schema,
		User:   url.UserPassword(smis.username, smis.password),
		Host:   smis.host + ":" + smis.port,
		Path:   "/ecom/edaa",
	}
}

/////////////////////////////////////////////////
// Parses URL into properly encoded URL format //
/////////////////////////////////////////////////

func parseURL(URL string) string {

	var newURL string

	for _, char := range URL {
		switch {
		case char == ' ':
			newURL = newURL + "%20"
		case char == '!':
			newURL = newURL + "%21"
		case char == '"':
			newURL = newURL + "%22"
		case char == '#':
			newURL = newURL + "%23"
		case char == '$':
			newURL = newURL + "%24"
		case char == '%':
			newURL = newURL + "%25"
		case char == '&':
			newURL = newURL + "%26"
		case char == '\'':
			newURL = newURL + "%27"
		case char == '(':
			newURL = newURL + "%28"
		case char == ')':
			newURL = newURL + "%29"
		case char == '*':
			newURL = newURL + "%2A"
		case char == '+':
			newURL = newURL + "%2B"
		case char == ',':
			newURL = newURL + "%2C"
		case char == '-':
			newURL = newURL + "%2D"
		case char == '.':
			newURL = newURL + "%2E"
		case char == ':':
			newURL = newURL + "%3A"
		case char == ';':
			newURL = newURL + "%3B"
		default:
			newURL = newURL + string(char)
		}
	}

	return newURL
}

func (smis *SMIS) query(httpType, objectPath string, body, resp interface{}) error {

	///////////////////////////////////////////
	//      Setup http/JSON header request   //
	///////////////////////////////////////////

	var URL string
	objectPath = parseURL(objectPath)

	if smis.insecure {
		URL = "http://" + smis.host + ":" + smis.port + objectPath
	} else {
		URL = "https://" + smis.host + ":" + smis.port + objectPath
	}

	// Create http request & add auth
	var req *http.Request

	if body != nil {
		// Parse out body struct into JSON
		bodyBytes, _ := json.Marshal(body)
		req, _ = http.NewRequest(httpType, URL, bytes.NewBuffer(bodyBytes))
	} else {
		req, _ = http.NewRequest(httpType, URL, nil)
	}

	// Create Authorization token which is then added to header of request. Auth token is in the format of username:password then encoded in base64.
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(smis.username + ":" + smis.password))

	// Add header specific items
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Host", smis.host+":"+smis.port)

	// Perform request
	httpResp, err := smis.client.Do(req)
	if err != nil {
		return err
	}

	// Cleanup Response
	defer httpResp.Body.Close()

	// Deal with errors
	switch {
	case httpResp.StatusCode == 400:
		return errors.New("JSON Build Error")
	case httpResp.StatusCode == 401:
		return errors.New("JSON Auth Error")
	case httpResp.StatusCode == 404:
		return errors.New("Object Could not be found")
	case httpResp == nil:
		return nil
	// Decode JSON of response into our interface defined for the specific request sent
	case httpResp.StatusCode == 200 || httpResp.StatusCode == 201:
		err = json.NewDecoder(httpResp.Body).Decode(resp)
		return err
	default:
		return errors.New("JSON Build Error")
	}
}

func (smis *SMIS) EnumerateInstanceNames(classname string) ([]gowbem.CIMInstanceName, error) {
	c, e := gowbem.NewClientCIMXML(getArrayUrl(smis), smis.insecure)
	if nil != e {
		return nil, e
	}
	return c.EnumerateInstanceNames(smis.namespace, classname)
}

func (smis *SMIS) EnumerateInstances(className string, deepInheritance bool, localOnly bool, includeQualifiers bool, includeClassOrigin bool, propertyList []string) ([]gowbem.CIMInstanceWithName, error) {
	c, e := gowbem.NewClientCIMXML(getArrayUrl(smis), smis.insecure)
	if nil != e {
		return nil, e
	}
	return c.EnumerateInstances(smis.namespace, className, deepInheritance, localOnly, includeQualifiers, includeClassOrigin, propertyList)
}

func (smis *SMIS) GetInstanceByInstanceName(instanceName gowbem.CIMInstanceName, propertyList []string) (gowbem.CIMInstance, error) {
	c, e := gowbem.NewClientCIMXML(getArrayUrl(smis), smis.insecure)
	if nil != e {
		return nil, e
	}
	return c.GetInstanceByInstanceName(smis.namespace, instanceName, true, true, true, propertyList)
}

func (smis *SMIS) AssociatorNames(instanceName gowbem.CIMInstanceName, assocClass, resultClass, role, resultRole string) ([]gowbem.CIMInstanceName, error) {
	c, e := gowbem.NewClientCIMXML(getArrayUrl(smis), smis.insecure)
	if nil != e {
		return nil, e
	}
	return c.AssociatorNames(smis.namespace, instanceName, assocClass, resultClass, role, resultRole)
}

func (smis *SMIS) AssociatorInstances(instanceName gowbem.CIMInstanceName, assocClass, resultClass, role, resultRole string, includeClassOrigin bool, propertyList []string) ([]gowbem.CIMInstance, error) {
	c, e := gowbem.NewClientCIMXML(getArrayUrl(smis), smis.insecure)
	if nil != e {
		return nil, e
	}
	return c.AssociatorInstances(smis.namespace, instanceName, assocClass, resultClass, role, resultRole, includeClassOrigin, propertyList)
}

func (smis *SMIS) ReferenceNames(instanceName gowbem.CIMInstanceName, resultClass, role string) ([]gowbem.CIMInstanceName, error) {
	c, e := gowbem.NewClientCIMXML(getArrayUrl(smis), smis.insecure)
	if nil != e {
		return nil, e
	}
	return c.ReferenceNames(smis.namespace, instanceName, resultClass, role)
}

func (smis *SMIS) EnumerateClassNames(className string, deep bool) ([]string, error) {
	c, e := gowbem.NewClientCIMXML(getArrayUrl(smis), smis.insecure)
	if nil != e {
		return nil, e
	}
	return c.EnumerateClassNames(smis.namespace, className, deep)
}

func (smis *SMIS) GetKeyFromInstanceName(instanceName gowbem.CIMInstanceName, key string) (interface{}, error) {
	for i := 0; i < instanceName.GetKeyBindings().Len(); i++ {
		if instanceName.GetKeyBindings().Get(i).GetName() == key {
			return instanceName.GetKeyBindings().Get(i).GetValue(), nil
		}
	}
	return "", errors.New("Key not found")
}
