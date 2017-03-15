package apiv1

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"GoWBEM/src/gowbem"
)

type SMIS struct {
	host     string
	port     string
	insecure bool
	username string
	password string
	client   *http.Client
	conn     *gowbem.WBEMConnection
}

func New(host string, port string, insecure bool, username string, password string) (*SMIS, error) {
	if host == "" || port == "" || username == "" || password == "" {
		return nil, errors.New("Missing host (SMIS Host IP), port (SMIS Host Port), username, or password \n Check Environment Variables..")
	}

	var client *http.Client
	if insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	return &SMIS{host, port, insecure, username, password, client, nil}, nil
}

func getArrayUrl(smis *SMIS) string {
	var schema string
	if smis.insecure {
		schema = "http"
	} else {
		schema = "https"
	}
	path := url.URL{
		Scheme: schema,
		User:   url.UserPassword(smis.username, smis.password),
		Host:   smis.host + ":" + smis.port,
		Path:   "/root/emc",
	}
	return path.String()
}

func GetWBEMConn(smis *SMIS) (*gowbem.WBEMConnection, error) {
	if smis.conn == nil {
		c, e := gowbem.NewWBEMConn(getArrayUrl(smis))
		if nil != e {
			return nil, e
		}
		smis.conn = c
	}
	return smis.conn, nil
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

func (smis *SMIS) EnumerateInstanceNames(classname string) ([]gowbem.InstanceName, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.EnumerateInstanceNames(MakeClassName(classname))
}

func (smis *SMIS) EnumerateInstances(className string, deepInheritance bool, includeClassOrigin bool, propertyList []string) ([]gowbem.ValueNamedInstance, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.EnumerateInstances(&gowbem.ClassName{Name: className}, deepInheritance, includeClassOrigin, propertyList)
}

func (smis *SMIS) GetInstance(instanceName *gowbem.InstanceName, includeClassOrigin bool, propertyList []string) (*gowbem.Instance, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	inst, err := c.GetInstance(instanceName, includeClassOrigin, propertyList)
	if err != nil {
		return nil, err
	}
	return &inst[0], err
}

func (smis *SMIS) AssociatorNames(instanceName *gowbem.InstanceName, assocClass, resultClass string, role, resultRole *string) ([]gowbem.ObjectPath, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.AssociatorNames(MakeObjectName(nil, instanceName), MakeClassName(assocClass), MakeClassName(resultClass), role, resultRole)
}

func (smis *SMIS) AssociatorInstances(instanceName *gowbem.InstanceName, assocClass, resultClass string, role, resultRole *string, includeClassOrigin bool, propertyList []string) ([]gowbem.ValueObjectWithPath, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.Associators(MakeObjectName(nil, instanceName), MakeClassName(assocClass), MakeClassName(resultClass), role, resultRole, includeClassOrigin, propertyList)
}

func (smis *SMIS) ReferenceNames(instanceName *gowbem.InstanceName, assocClass string, role *string) ([]gowbem.ObjectPath, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.ReferenceNames(MakeObjectName(nil, instanceName), MakeClassName(assocClass), role)
}

func (smis *SMIS) EnumerateClassNames(className string, deepInheritance bool) ([]gowbem.Class, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.EnumerateClassNames(MakeClassName(className), deepInheritance)
}

func (smis *SMIS) InvokeMethod(instanceName *gowbem.InstanceName, methodName string, paramValues []gowbem.IParamValue) (int, []gowbem.ParamValue, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return -1, nil, e
	}

	return c.InvokeMethod(MakeObjectName(nil, instanceName), methodName, paramValues)
}

func (smis *SMIS) GetKeyFromInstanceName(instanceName *gowbem.InstanceName, keyName string) (interface{}, error) {
	for _, key := range instanceName.KeyBinding {
		if keyName == key.Name {
			return key.KeyValue.KeyValue, nil
		}
	}
	return "", errors.New("Key not found")
}

func (smis *SMIS) GetPropertyByName(instance *gowbem.Instance, name string) (interface{}, error) {
	for _, pr := range instance.Property {
		if pr.Name == name {
			return pr.Value.Value, nil
		}
	}
	return "", errors.New("Property not found")
}

func MakeClassName(name string) *gowbem.ClassName {
	return &gowbem.ClassName{Name: name}
}

func MakeObjectName(classname *gowbem.ClassName, instanceName *gowbem.InstanceName) *gowbem.ObjectName {
	return &gowbem.ObjectName{ClassName: classname, InstanceName: instanceName}
}
