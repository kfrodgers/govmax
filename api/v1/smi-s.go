package apiv1

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"

	"github.com/kfrodgers/GoWBEM/src/gowbem"
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

/////////////////
// getArrayUrl //
/////////////////

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

/////////////////
// GetWBEMConn //
/////////////////

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

////////////////////////////
// EnumerateInstanceNames //
////////////////////////////

func (smis *SMIS) EnumerateInstanceNames(classname string) ([]gowbem.InstanceName, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.EnumerateInstanceNames(MakeClassName(classname))
}

////////////////////////
// EnumerateInstances //
////////////////////////

func (smis *SMIS) EnumerateInstances(className string, deepInheritance bool, includeClassOrigin bool, propertyList []string) ([]gowbem.ValueNamedInstance, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.EnumerateInstances(&gowbem.ClassName{Name: className}, deepInheritance, includeClassOrigin, propertyList)
}

/////////////////
// GetInstance //
/////////////////

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

/////////////////////
// AssociatorNames //
/////////////////////

func (smis *SMIS) AssociatorNames(instanceName *gowbem.InstanceName, assocClass, resultClass string, role, resultRole *string) ([]gowbem.ObjectPath, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.AssociatorNames(MakeObjectName(nil, instanceName), MakeClassName(assocClass), MakeClassName(resultClass), role, resultRole)
}

/////////////////////////
// AssociatorInstances //
/////////////////////////

func (smis *SMIS) AssociatorInstances(instanceName *gowbem.InstanceName, assocClass, resultClass string, role, resultRole *string, includeClassOrigin bool, propertyList []string) ([]gowbem.ValueObjectWithPath, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.Associators(MakeObjectName(nil, instanceName), MakeClassName(assocClass), MakeClassName(resultClass), role, resultRole, includeClassOrigin, propertyList)
}

////////////////////
// ReferenceNames //
////////////////////

func (smis *SMIS) ReferenceNames(instanceName *gowbem.InstanceName, assocClass string, role *string) ([]gowbem.ObjectPath, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.ReferenceNames(MakeObjectName(nil, instanceName), MakeClassName(assocClass), role)
}

/////////////////////////
// EnumerateClassNames //
/////////////////////////

func (smis *SMIS) EnumerateClassNames(className string, deepInheritance bool) ([]gowbem.Class, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return nil, e
	}
	return c.EnumerateClassNames(MakeClassName(className), deepInheritance)
}

//////////////////
// InvokeMethod //
//////////////////

func (smis *SMIS) InvokeMethod(instanceName *gowbem.InstanceName, methodName string, paramValues []gowbem.IParamValue) (int, []gowbem.ParamValue, error) {
	c, e := GetWBEMConn(smis)
	if nil != e {
		return -1, nil, e
	}

	return c.InvokeMethod(MakeObjectName(nil, instanceName), methodName, paramValues)
}

//////////////////
// InvokeMethod //
//////////////////

func GetKeyFromInstanceName(instanceName *gowbem.InstanceName, keyName string) (interface{}, error) {
	for _, key := range instanceName.KeyBinding {
		if keyName == key.Name {
			return key.KeyValue.KeyValue, nil
		}
	}
	return "", errors.New("Key not found")
}

///////////////////////
// GetPropertyByName //
///////////////////////

func GetPropertyByName(instance *gowbem.Instance, name string) (interface{}, error) {
	for _, pr := range instance.Property {
		if pr.Name == name {
			return pr.Value.Value, nil
		}
	}
	return "", errors.New("Property not found")
}

///////////////////
// MakeClassName //
///////////////////

func MakeClassName(name string) *gowbem.ClassName {
	return &gowbem.ClassName{Name: name}
}

////////////////////
// MakeObjectName //
////////////////////

func MakeObjectName(classname *gowbem.ClassName, instanceName *gowbem.InstanceName) *gowbem.ObjectName {
	return &gowbem.ObjectName{ClassName: classname, InstanceName: instanceName}
}
