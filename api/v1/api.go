package apiv1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/runner-mei/gowbem"
)

///////////////////////////////////////////////////////////////
//            GET a list of Storage Arrays                   //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageArrays() ([]gowbem.CIMInstanceName, error) {
	return smis.EnumerateInstanceNames("Symm_StorageSystem")
}

func (smis *SMIS) GetStorageInstanceName(sid string) (gowbem.CIMInstanceName, error) {
	arrays, err := smis.GetStorageArrays()
	if err != nil {
		return nil, err
	}
	for _, array := range arrays {
		name, err := smis.GetKeyFromInstanceName(array, "Name")
		if err != nil {
			continue
		}
		if strings.HasSuffix(name.(string), sid) {
			return array, nil
		}
	}
	return nil, errors.New("Array not found")
}

func (smis *SMIS) GetStorageConfigurationService(systemInstanceName gowbem.CIMInstanceName) (gowbem.CIMInstanceName, error) {
	configServices, err := smis.EnumerateInstanceNames("EMC_StorageConfigurationService")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")
	for _, service := range configServices {
		name, _ := smis.GetKeyFromInstanceName(service, "SystemName")
		if name.(string) == sysName.(string) {
			return service, nil
		}
	}
	return nil, errors.New("Service not found")
}

func (smis *SMIS) GetControllerConfigurationService(systemInstanceName gowbem.CIMInstanceName) (gowbem.CIMInstanceName, error) {
	controllerServices, err := smis.EnumerateInstanceNames("EMC_ControllerConfigurationService")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")
	for _, service := range controllerServices {
		name, _ := smis.GetKeyFromInstanceName(service, "SystemName")
		if name.(string) == sysName.(string) {
			return service, nil
		}
	}
	return nil, errors.New("Service not found")
}

func (smis *SMIS) GetStorageHardwareIDManagementService(systemInstanceName gowbem.CIMInstanceName) (gowbem.CIMInstanceName, error) {
	managementServices, err := smis.EnumerateInstanceNames("Symm_StorageHardwareIDManagementService")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")
	for _, service := range managementServices {
		name, _ := smis.GetKeyFromInstanceName(service, "SystemName")
		if name.(string) == sysName.(string) {
			return service, nil
		}
	}
	return nil, errors.New("Service not found")
}

func (smis *SMIS) GetSoftwareIdentity(systemInstanceName gowbem.CIMInstanceName) (gowbem.CIMInstance, error) {
	softwareIdents, err := smis.EnumerateInstanceNames("Symm_StorageSystemSoftwareIdentity")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")
	for _, swIdent := range softwareIdents {
		name, _ := smis.GetKeyFromInstanceName(swIdent, "InstanceID")
		if name.(string) == sysName.(string) {
			return smis.GetInstanceByInstanceName(swIdent, nil)
		}
	}
	return nil, errors.New("Service not found")
}

func (smis *SMIS) IsArrayV3(systemInstanceName gowbem.CIMInstanceName) bool {
	swIdent, err := smis.GetSoftwareIdentity(systemInstanceName)
	if err != nil {
		return false
	}
	property := swIdent.GetPropertyByName("EMCEnginuityFamily")
	ucode, _ := strconv.Atoi(property.GetValue().(string))
	return ucode >= 5900
}

///////////////////////////////////////////////////////////////
//            GET a list of Storage Pools                    //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStoragePools(systemInstanceName gowbem.CIMInstanceName) ([]gowbem.CIMInstanceName, error) {
	if smis.IsArrayV3(systemInstanceName) {
		return smis.AssociatorNames(systemInstanceName, "", "Symm_SRPStoragePool", "", "")
	} else {
		return smis.AssociatorNames(systemInstanceName, "", "EMC_VirtualProvisioningPool", "", "")
	}
}

///////////////////////////////////////////////////////////////
//            GET a list of Masking Views                    //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetMaskingViews(systemInstanceName gowbem.CIMInstanceName) ([]gowbem.CIMInstanceName, error) {
	return smis.AssociatorNames(systemInstanceName, "", "Symm_LunMaskingView", "", "")
}

///////////////////////////////////////////////////////////////
//         GET a list of Storage (Device) Groups             //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetStorageGroups(systemInstanceName gowbem.CIMInstanceName) ([]gowbem.CIMInstanceName, error) {
	controllerService, err := smis.GetControllerConfigurationService(systemInstanceName)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(controllerService, "", "SE_DeviceMaskingGroup", "", "")
}

///////////////////////////////////////////////////////////////
//         GET a list of Port (Target) Groups                //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetPortGroups(systemInstanceName gowbem.CIMInstanceName) ([]gowbem.CIMInstanceName, error) {
	controllerService, err := smis.GetControllerConfigurationService(systemInstanceName)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(controllerService, "", "SE_TargetMaskingGroup", "", "")
}

///////////////////////////////////////////////////////////////
//         GET a list of Host (Initiator) Groups             //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetHostGroups(systemInstanceName gowbem.CIMInstanceName) ([]gowbem.CIMInstanceName, error) {
	controllerService, err := smis.GetControllerConfigurationService(systemInstanceName)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(controllerService, "", "SE_InitiatorMaskingGroup", "", "")
}

///////////////////////////////////////////////////////////////
//            GET a list of Storage Volumes                  //
///////////////////////////////////////////////////////////////

func (smis *SMIS) GetVolumes(systemInstance gowbem.CIMInstanceName) ([]gowbem.CIMInstanceName, error) {
	return smis.AssociatorNames(systemInstance, "", "CIM_StorageVolume", "", "")
}

///////////////////////////////////////////////////////////
//            GET a Storage Volume by ID                 //
///////////////////////////////////////////////////////////

func (smis *SMIS) GetVolumeByID(systemInstance gowbem.CIMInstanceName, volumeID string) (*gowbem.CIMInstanceName, error) {
	volumes, err := smis.GetVolumes(systemInstance)
	if err != nil {
		return nil, err
	}
	for _, volume := range volumes {
		name, err := smis.GetKeyFromInstanceName(volume, "DeviceID")
		if err == nil {
			if name.(string) == volumeID {
				return &volume, nil
			}
		}
	}
	return nil, errors.New("Volume not found")
}

///////////////////////////////////////////////////////////
//            GET a Storage Volume by Name               //
///////////////////////////////////////////////////////////

func (smis *SMIS) GetVolumeByName(systemInstance gowbem.CIMInstanceName, volumeName string) (*gowbem.CIMInstanceName, error) {
	volumes, err := smis.GetVolumes(systemInstance)
	if err != nil {
		return nil, err
	}

	var volumeInstance gowbem.CIMInstance
	for _, volume := range volumes {
		volumeInstance, err = smis.GetInstanceByInstanceName(volume, nil)
		if err == nil {
			nameProp := volumeInstance.GetPropertyByName("ElementName")
			if nameProp.GetValue().(string) == volumeName {
				return &volume, nil
			}
		}
	}
	return nil, errors.New("Volume not found")
}

//////////////////////////////////////////////////////////////////////////////////////////////////
//      GET a Job Status                                                                        //
//                                                                                              //
//  2 - New: job has not been started                                                           //
//  3 - Starting: job is moving into running state                                              //
//  4 - Running: job is running                                                                 //
//  5 - Suspended: job is stopped, but can be restarted                                         //
//  6 - Shutting Down: job is moving to an completed/terminated/killed state                    //
//  7 - Completed: job has been completed normally                                              //
//  8 - Terminated: job has been stopped by a terminate state change request                    //
//  9 - Killed: job has stopped by a kill state change request                                  //
//  10 - Exception: job is in an abnormal state due to an error condition                       //
//  11 - Service: job is in a vendor-specific state that supports problem discovery/resolution  //
//  12 - Query Pending: job is waiting for a client to resolve a query                          //
//////////////////////////////////////////////////////////////////////////////////////////////////

func (smis *SMIS) GetJobStatus(jobName gowbem.CIMInstanceName) (resp *gowbem.CIMInstance, JobStatus string, err error) {
	*resp, err = smis.GetInstanceByInstanceName(jobName, nil)
	if err != nil {
		return nil, "", err
	}
	jobStatusMap := map[int]string{
		2:  "NEW",
		3:  "STARTING",
		4:  "RUNNING",
		5:  "SUSPENDED",
		6:  "SHUTTING_DOWN",
		7:  "COMPLETED",
		8:  "TERMINATED",
		9:  "KILLED",
		10: "EXCEPTION",
		11: "SERVICE",
		12: "QUERY_PENDING",
	}

	var jobState int
	var jobStatus string
	var ok bool
	jobState, _ = strconv.Atoi((*resp).GetPropertyByName("JobState").GetValue().(string))
	if jobStatus, ok = jobStatusMap[jobState]; !ok {
		jobStatus = "UNKNOWN"
	}
	return resp, jobStatus, err
}

//////////////////////////////////////
//    REQUEST Structs used for      //
//   volume creation on the VMAX3.  //
//////////////////////////////////////

type PostVolumesReq struct {
	PostVolumesRequestContent *PostVolumesReqContent `json:"content"`
}

type PostVolumesReqContent struct {
	AtType             string `json:"@type"`
	ElementName        string `json:"ElementName"`
	ElementType        string `json:"ElementType"`
	EMCNumberOfDevices string `json:"EMCNumberOfDevices"`
	Size               string `json:"Size"`
}

////////////////////////////////////////////////////////////
//            RESPONSE Struct used for                    //
//          volume creation on the VMAX3.                 //
////////////////////////////////////////////////////////////

type PostVolumesResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
				I_Size int `json:"i$Size"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

///////////////////////////////////////////////////////////
//              CREATE a Storage Volume                  //
//     and check for Volume Creation Completion          //
///////////////////////////////////////////////////////////

func (smis *SMIS) PostVolumes(req *PostVolumesReq, sid string) (resp1 *PostVolumesResp, resp2 *gowbem.CIMInstance, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_StorageConfigurationService/CreationClassName::Symm_StorageConfigurationService,Name::EMCStorageConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/CreateOrModifyElementFromStoragePool", req, &resp1)

	err = smis.query("GET", "/ecom/edaa/root/emc/instances/SE_ConcreteJob/InstanceID::"+resp1.Entries[0].Content.I_Parameters.I_Job.E0_InstanceID, nil, &resp2)
	JobStatus := 7
	for JobStatus == 2 || JobStatus == 3 || JobStatus == 4 || JobStatus == 5 || JobStatus == 6 || JobStatus == 12 {
		err = smis.query("GET", "/ecom/edaa/root/emc/instances/SE_ConcreteJob/InstanceID::"+resp1.Entries[0].Content.I_Parameters.I_Job.E0_InstanceID, nil, &resp2)
		JobStatus = 7
	}
	if JobStatus != 7 {
		fmt.Println("Error: Volume creation incomplete")
	}
	return resp1, resp2, err
}

//////////////////////////////////////
//   REQUEST Structs used for any   //
//   group creation on the VMAX3.   //
//                                  //
//    Storage Group (SG) - Type 4   //
//     Port Group (PG) - Type 3     //
//   Initiator Group (IG) - Type 2  //
//////////////////////////////////////

type PostGroupReq struct {
	PostGroupRequestContent *PostGroupReqContent `json:"content"`
}

type PostGroupReqContent struct {
	AtType    string `json:"@type"`
	GroupName string `json:"GroupName"`
	Type      string `json:"Type"`
}

////////////////////////////////////////////////////////////
//           RESPONSE Struct  used for any                //
//           group creation on the VMAX3.                 //
//                                                        //
//   Storage Group (SG) - Type SE_DeviceMaskingGroup      //
//      Port Group (PG) - Type SE_TargetMaskingGroup      //
//  Initiator Group (IG) - Type SE_InitiatorMaskingGroup  //
////////////////////////////////////////////////////////////

type PostGroupResp struct {
	Entries []struct {
		Content struct {
			_type        string `json:"@type"`
			I_parameters struct {
				I_MaskingGroup struct {
					_type         string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$MaskingGroup"`
			} `json:"i$parameters"`
			I_returnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

///////////////////////////////////////////////////////////////
//                  CREATE an Array Group                    //
// Type Depends on Type field specified in requesting struct //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostCreateGroup(req *PostGroupReq, sid string) (resp *PostGroupResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::SYMMETRIX-+-"+sid+"/action/CreateGroup", req, &resp)
	return resp, err
}

////////////////////////////////////////////////////////////
//      RESPONSE Struct used for each SRP settings        //
////////////////////////////////////////////////////////////

type GetStoragePoolSettingsResp struct {
	Entries []struct {
		Content struct {
			AtType                           string  `json:"@type"`
			I_Changeable                     bool    `json:"i$Changeable"`
			I_ChangeableType                 int     `json:"i$ChangeableType"`
			I_CompressedElement              bool    `json:"i$CompressedElement"`
			I_CompressionRate                int     `json:"i$CompressionRate"`
			I_DataRedundancyGoal             int     `json:"i$DataRedundancyGoal"`
			I_DataRedundancyMax              int     `json:"i$DataRedundancyMax"`
			I_DataRedundancyMin              int     `json:"i$DataRedundancyMin"`
			I_DeltaReservationGoal           int     `json:"i$DeltaReservationGoal"`
			I_DeltaReservationMax            int     `json:"i$DeltaReservationMax"`
			I_DeltaReservationMin            int     `json:"i$DeltaReservationMin"`
			I_ElementName                    string  `json:"i$ElementName"`
			I_EMCApproxAverageResponseTime   float64 `json:"i$EMCApproxAverageResponseTime"`
			I_EMCDeduplicationRate           int     `json:"i$EMCDeduplicationRate"`
			I_EMCEnableDIF                   int     `json:"i$EMCEnableDIF"`
			I_EMCEnableEFDCache              int     `json:"i$EMCEnableEFDCache"`
			I_EMCFastSetting                 string  `json:"i$EMCFastSetting"`
			I_EMCParticipateInPowerSavings   int     `json:"i$EMCParticipateInPowerSavings"`
			I_EMCPoolCompressionState        int     `json:"i$EMCPoolCompressionState"`
			I_EMCPottedSetting               bool    `json:"i$EMCPottedSetting"`
			I_EMCRaidGroupLUN                bool    `json:"i$EMCRaidGroupLUN"`
			I_EMCRaidLevel                   string  `json:"i$EMCRaidLevel"`
			I_EMCSLO                         string  `json:"i$EMCSLO"`
			I_EMCSLOBaseName                 string  `json:"i$EMCSLOBaseName"`
			I_EMCSLOdescription              string  `json:"i$EMCSLOdescription"`
			I_EMCSRP                         string  `json:"i$EMCSRP"`
			I_EMCStorageSettingType          int     `json:"i$EMCStorageSettingType"`
			I_EMCUniqueID                    string  `json:"i$EMCUniqueID"`
			I_EMCWorkload                    string  `json:"i$EMCWorkload"`
			I_ExtentStripeLength             int     `json:"i$ExtentStripeLength"`
			I_ExtentStripeLengthMax          int     `json:"i$ExtentStripeLengthMax"`
			I_ExtentStripeLengthMin          int     `json:"i$ExtentStripeLengthMin"`
			I_InitialStorageTierMethodology  int     `json:"i$InitialStorageTierMethodology"`
			I_InitialStorageTieringSelection int     `json:"i$InitialStorageTieringSelection"`
			I_InitialSynchronization         int     `json:"i$InitialSynchronization"`
			I_InstanceID                     string  `json:"i$InstanceID"`
			I_NoSinglePointOfFailure         bool    `json:"i$NoSinglePointOfFailure"`
			I_PackageRedundancyGoal          int     `json:"i$PackageRedundancyGoal"`
			I_PackageRedundancyMax           int     `json:"i$PackageRedundancyMax"`
			I_PackageRedundancyMin           int     `json:"i$PackageRedundancyMin"`
			I_SpaceLimit                     int     `json:"i$SpaceLimit"`
			I_StorageExtentInitialUsage      int     `json:"i$StorageExtentInitialUsage"`
			I_StoragePoolInitialUsage        int     `json:"i$StoragePoolInitialUsage"`
			I_ThinProvisionedPoolType        int     `json:"i$ThinProvisionedPoolType"`
			I_UseReplicationBuffer           int     `json:"i$UseReplicationBuffer"`
			Links                            []struct {
				Href string `json:"href"`
				Rel  string `json:"rel"`
			} `json:"links"`
			Xmlns_i string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Gd_etag      string `json:"gd$etag"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

///////////////////////////////////////////////////////////////
//                GET Storage Pool Settings                  //
///////////////////////////////////////////////////////////////
func (smis *SMIS) GetStoragePoolCapabilities(srp_name gowbem.CIMInstanceName) (gowbem.CIMInstanceName, error) {
	capabilities, err := smis.EnumerateInstanceNames("Symm_StoragePoolCapabilities")

	name, err := smis.GetKeyFromInstanceName(srp_name, "InstanceID")
	if err != nil {
		return nil, err
	}
	for _, entry := range capabilities {
		key, err := smis.GetKeyFromInstanceName(entry, "InstanceID")
		if err != nil {
			continue
		}
		if key.(string) == name.(string) {
			return entry, nil
		}
	}
	return nil, errors.New("Capabilities not found")

}

func (smis *SMIS) GetStoragePoolSettings(srp_name gowbem.CIMInstanceName) ([]gowbem.CIMInstanceName, error) {
	capabilities, err := smis.GetStoragePoolCapabilities(srp_name)
	if err != nil {
		return nil, err
	}
	return smis.AssociatorNames(capabilities, "", "CIM_StorageSetting", "", "")
}

///////////////////////////////////////////////////////////////
//        Struct used to store all SLO information           //
///////////////////////////////////////////////////////////////

type SLO_Struct struct {
	SLO_Name    string
	respTime    float64
	SRP         string
	Workload    string
	ElementName string
	InstanceID  string
}

////////////////////////////////////////////////////////////////
//               GET Storage Level Objectives                 //
//                                                            //
//             1 -> Get Storage Pools of VMAX3                //
//    2 -> Get Storage Pool Settings of each Storage Pool     //
//   3 -> Parse out SLO information of VMAX3 and return it    //
////////////////////////////////////////////////////////////////

func (smis *SMIS) GetSLOs(systemInstanceName gowbem.CIMInstanceName) (SLOs []SLO_Struct, err error) {
	if !smis.IsArrayV3(systemInstanceName) {
		return nil, errors.New("SLOs not supportted")
	}

	storagePools, err := smis.GetStoragePools(systemInstanceName)
	if err != nil {
		return nil, err
	}

	for _, SRP := range storagePools {
		storagePoolSettings, err := smis.GetStoragePoolSettings(SRP)
		if err != nil {
			return nil, err
		}
		for _, storagePoolSetting := range storagePoolSettings {
			base_name, _ := smis.GetKeyFromInstanceName(storagePoolSetting, "EMCSLOBaseName")
			resp_time, _ := smis.GetKeyFromInstanceName(storagePoolSetting, "EMCApproxAverageResponseTime")
			srp, _ := smis.GetKeyFromInstanceName(storagePoolSetting, "EMCSRP")
			workload, _ := smis.GetKeyFromInstanceName(storagePoolSetting, "EMCWorkload")
			elem_name, _ := smis.GetKeyFromInstanceName(storagePoolSetting, "ElementName")
			inst_id, _ := smis.GetKeyFromInstanceName(storagePoolSetting, "InstanceID")
			newSLO := SLO_Struct{
				SLO_Name:    base_name.(string),
				respTime:    resp_time.(float64),
				SRP:         srp.(string),
				Workload:    workload.(string),
				ElementName: elem_name.(string),
				InstanceID:  inst_id.(string),
			}
			SLOs = append(SLOs, newSLO)
		}
	}
	return SLOs, nil
}

/////////////////////////////////////////////////////////
//               REQUEST Structs used for              //
//         adding AND removing a volume to/from        //
//             a storage group on the VMAX3.           //
/////////////////////////////////////////////////////////

type PostVolumesToSGReq struct {
	PostVolumesToSGRequestContent *PostVolumesToSGReqContent `json:"content"`
}

type PostVolumesToSGReqContent struct {
	AtType                              string                             `json:"@type"`
	PostVolumesToSGRequestContentMG     *PostVolumesToSGReqContentMG       `json:"MaskingGroup"`
	PostVolumesToSGRequestContentMember []*PostVolumesToSGReqContentMember `json:"Members"`
}

type PostVolumesToSGReqContentMG struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

type PostVolumesToSGReqContentMember struct {
	AtType                  string `json:"@type"`
	CreationClassName       string `json:"CreationClassName"`
	DeviceID                string `json:"DeviceID"`
	SystemCreationClassName string `json:"SystemCreationClassName"`
	SystemName              string `json:"SystemName"`
}

////////////////////////////////////////////////////////////
//               RESPONSE Struct used for                 //
//         adding AND removing a volume to/from           //
//             a storage group on the VMAX3.              //
////////////////////////////////////////////////////////////

type PostVolumesToSGResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

///////////////////////////////////////////////////////////////
//             ADD Volumes to a Storage Group                //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostVolumesToSG(req *PostVolumesToSGReq, sid string) (resp *PostVolumesToSGResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/AddMembers", req, &resp)
	return resp, err
}

///////////////////////////////////////////////////////////////
//          REMOVE Volumes from a Storage Group              //
///////////////////////////////////////////////////////////////

func (smis *SMIS) RemoveVolumeFromSG(req *PostVolumesToSGReq, sid string) (resp *PostVolumesToSGResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/RemoveMembers", req, &resp)
	return resp, err
}

/////////////////////////////////////////////////////////
//               REQUEST Structs used for              //
//   creating a storage hardware ID for an initiator   //
/////////////////////////////////////////////////////////

type PostStorageHardwareIDReq struct {
	PostStorageHardwareIDRequestContent *PostStorageHardwareIDReqContent `json:"content"`
}

type PostStorageHardwareIDReqContent struct {
	AtType    string `json:"@type"`
	IDType    string `json:"IDType"`
	StorageID string `json:"StorageID"`
}

////////////////////////////////////////////////////////////
//            RESPONSE Struct used for                    //
//   creating a storage hardware ID for an initiator      //
////////////////////////////////////////////////////////////

type PostStorageHardwareIDResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_HardwareID struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$HardwareID"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

/////////////////////////////////////////////////////////
//               REQUEST Struct used for               //
//       adding AND removing an initiator to/from      //
//              a host group on the VMAX3.             //
/////////////////////////////////////////////////////////

type PostInitiatorToHGReq struct {
	PostInitiatorToHGRequestContent *PostInitiatorToHGReqContent `json:"content"`
}

type PostInitiatorToHGReqContent struct {
	AtType                                string                               `json:"@type"`
	PostInitiatorToHGRequestContentMG     *PostInitiatorToHGReqContentMG       `json:"MaskingGroup"`
	PostInitiatorToHGRequestContentMember []*PostInitiatorToHGReqContentMember `json:"Members"`
}

type PostInitiatorToHGReqContentMG struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

type PostInitiatorToHGReqContentMember struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

////////////////////////////////////////////////////////////
//                RESPONSE Struct used for                //
//        adding AND removing an initiator to/from        //
//             a host group on the VMAX3.                 //
////////////////////////////////////////////////////////////

type PostInitiatorToHGResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

///////////////////////////////////////////////////////////////
//             ADD Initiators to a Host Group                //
//                                                           //
//     1 -> Create Storage Hardware ID for the Initiator     //
//     2 -> Add Storage Hardware ID to Initiator Group       //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostStorageHardwareID(req *PostStorageHardwareIDReq, sid string) (resp *PostStorageHardwareIDResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_StorageHardwareIDManagementService/CreationClassName::Symm_StorageHardwareIDManagementService,Name::EMCStorageHardwareIDManagementService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/CreateStorageHardwareID", req, &resp)
	return resp, err
}

func (smis *SMIS) PostInitiatorToHG(req *PostInitiatorToHGReq, sid string) (resp *PostInitiatorToHGResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/AddMembers", req, &resp)
	return resp, err
}

///////////////////////////////////////////////////////////////
//          REMOVE Initiators from a Host Group              //
//     (Requires a Storage Hardware ID from the Initiator)   //
///////////////////////////////////////////////////////////////

func (smis *SMIS) RemoveInitiatorFromHG(req *PostInitiatorToHGReq, sid string) (resp *PostInitiatorToHGResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/RemoveMembers", req, &resp)
	return resp, err
}

/////////////////////////////////////////////////////////
//               REQUEST Structs used for              //
//         adding AND removing a port to/from          //
//           a port group on the VMAX3.                //
/////////////////////////////////////////////////////////

type PostPortToPGReq struct {
	PostPortToPGRequestContent *PostPortToPGReqContent `json:"content"`
}

type PostPortToPGReqContent struct {
	AtType                           string                          `json:"@type"`
	PostPortToPGRequestContentMG     *PostPortToPGReqContentMG       `json:"MaskingGroup"`
	PostPortToPGRequestContentMember []*PostPortToPGReqContentMember `json:"Members"`
}

type PostPortToPGReqContentMG struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

type PostPortToPGReqContentMember struct {
	AtType                  string `json:"@type"`
	CreationClassName       string `json:"CreationClassName"`
	Name                    string `json:"Name"`
	SystemCreationClassName string `json:"SystemCreationClassName"`
	SystemName              string `json:"SystemName"`
}

///////////////////////////////////////////////////////
//            RESPONSE Struct used for               //
//       adding AND removing a port to/from          //
//            a port group on the VMAX3.             //
///////////////////////////////////////////////////////

type PostPortToPGResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

/////////////////////////////////////////////////////////////////////
//                     ADD Ports to a Port Group                   //
//                                                                 //
//    1 -> GET a list of Available Interfaces (aka FE Directors)   //
// 2 -> GET a list of Front End Adapter Endpoints (aka FE Ports)   //
//                  3 -> ADD Ports to Port Groups                  //
/////////////////////////////////////////////////////////////////////

func (smis *SMIS) PostPortToPG(req *PostPortToPGReq, sid string) (resp *PostPortToPGResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/AddMembers", req, &resp)
	return resp, err
}

///////////////////////////////////////////////////////////////
//             REMOVE Ports from a Port Group                //
///////////////////////////////////////////////////////////////

func (smis *SMIS) RemovePortFromPG(req *PostPortToPGReq, sid string) (resp *PostPortToPGResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/RemoveMembers", req, &resp)
	return resp, err
}

/////////////////////////////////////////////////////////
//               REQUEST Structs used for              //
//        creating a masking view on the VMAX3.        //
/////////////////////////////////////////////////////////

type PostCreateMaskingViewReq struct {
	PostCreateMaskingViewRequestContent *PostCreateMaskingViewReqContent `json:"content"`
}

type PostCreateMaskingViewReqContent struct {
	AtType                           string                        `json:"@type"`
	ElementName                      string                        `json:"ElementName"`
	PostInitiatorMaskingGroupRequest *PostInitiatorMaskingGroupReq `json:"InitiatorMaskingGroup"`
	PostTargetMaskingGroupRequest    *PostTargetMaskingGroupReq    `json:"TargetMaskingGroup"`
	PostDeviceMaskingGroupRequest    *PostDeviceMaskingGroupReq    `json:"DeviceMaskingGroup"`
}

type PostInitiatorMaskingGroupReq struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}
type PostTargetMaskingGroupReq struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}
type PostDeviceMaskingGroupReq struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

////////////////////////////////////////////////////////////
//            RESPONSE Struct used for                    //
//        creating a masking view on the VMAX3.           //
////////////////////////////////////////////////////////////

type PostCreateMaskingViewResp struct {
	Xmlns_gd string `json:"xmlns$gd"`
	Updated  string `json:"updated"`
	ID       string `json:"id"`

	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`

	Entries []struct {
		Updated string `json:"updated"`

		Links []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`

		Content_type string `json:"content-type"`

		Content struct {
			AtType       string `json:"@type"`
			Xmlns_i      string `json:"xmlns$i"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					Xmlns_e0      string `json:"xmlns$e0"`
					E0_InstanceID string `json:"e0$InstanceID"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_ReturnValue int `json:"i$returnValue"`
		} `json:"content"`
	} `json:"entries"`
}

///////////////////////////////////////////////////////////////
//                  CREATE a Masking View                    //
///////////////////////////////////////////////////////////////

func (smis *SMIS) PostCreateMaskingView(req *PostCreateMaskingViewReq, sid string) (resp *PostCreateMaskingViewResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/CreateMaskingView", req, &resp)
	return resp, err
}

////////////////////////////////////////////////////////////////
//     Struct used to store all Baremetal HBA Information     //
////////////////////////////////////////////////////////////////

type HostAdapter struct {
	HostID string
	WWN    string
}

////////////////////////////////////////////////////////////////
//             GET Baremetal HBA Information                  //
////////////////////////////////////////////////////////////////

func GetBaremetalHBA() (myHosts []HostAdapter, err error) {
	//Works for RedHat 5 and above (including CentOS and SUSE Linux)
	hostDir, _ := ioutil.ReadDir("/sys/class/scsi_host/")

	for _, host := range hostDir {
		if _, err := os.Stat("/sys/class/scsi_host/" + host.Name() + "/device/fc_host/" + host.Name() + "/port_name"); err == nil {
			byteArray, _ := ioutil.ReadFile("/sys/class/scsi_host/" + host.Name() + "/device/fc_host/" + host.Name() + "/port_name")
			newHost := HostAdapter{
				HostID: host.Name(),
				WWN:    string(byteArray),
			}
			myHosts = append(myHosts, newHost)
		}
	}
	return myHosts, nil
}

//////////////////////////////////////////////////////////////
//             REQUEST Structs used for any                 //
//             group deletion on the VMAX3.                 //
//                                                          //
//   Storage Group (SG) - AtType SE_DeviceMaskingGroup      //
//      Port Group (PG) - AtType SE_TargetMaskingGroup      //
//  Initiator Group (IG) - AtType SE_InitiatorMaskingGroup  //
//////////////////////////////////////////////////////////////

type DeleteGroupReq struct {
	DeleteGroupRequestContent *DeleteGroupReqContent `json:"content"`
}

type DeleteGroupReqContent struct {
	AtType                                string                             `json:"@type"`
	DeleteGroupRequestContentMaskingGroup *DeleteGroupReqContentMaskingGroup `json:"MaskingGroup"`
}

type DeleteGroupReqContentMaskingGroup struct {
	AtType     string `json:"@type"`
	InstanceID string `json:"InstanceID"`
}

////////////////////////////////////////////////////////////
//           RESPONSE Struct used for any                 //
//           group deletion on the VMAX3.                 //
////////////////////////////////////////////////////////////

type DeleteGroupResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_returnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

/////////////////////////////////////////////////////////////////
//                  DELETE an Array Group                      //
// Type Depends on AtType field specified in requesting struct //
/////////////////////////////////////////////////////////////////

func (smis *SMIS) PostDeleteGroup(req *DeleteGroupReq, sid string) (resp *DeleteGroupResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/DeleteGroup", req, &resp)
	return resp, err
}

//////////////////////////////////////////////////////////////
//             REQUEST Structs used for any                 //
//             volume deletion on the VMAX3.                //
//////////////////////////////////////////////////////////////

type DeleteVolReq struct {
	DeleteVolRequestContent *DeleteVolReqContent `json:"content"`
}

type DeleteVolReqContent struct {
	AtType                         string                      `json:"@type"`
	DeleteVolRequestContentElement *DeleteVolReqContentElement `json:"TheElement"`
}

type DeleteVolReqContentElement struct {
	AtType                  string `json:"@type"`
	DeviceID                string `json:"DeviceID"`
	CreationClassName       string `json:"CreationClassName"`
	SystemName              string `json:"SystemName"`
	SystemCreationClassName string `json:"SystemCreationClassName"`
}

////////////////////////////////////////////////////////////
//           RESPONSE Struct used for any                 //
//           volume deletion on the VMAX3.                //
////////////////////////////////////////////////////////////

type DeleteVolResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_Parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_ReturnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

/////////////////////////////////////////////////////////////////
//                  DELETE a Volume                            //
/////////////////////////////////////////////////////////////////

func (smis *SMIS) PostDeleteVol(req *DeleteVolReq, sid string) (resp *DeleteVolResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_StorageConfigurationService/CreationClassName::Symm_StorageConfigurationService,Name::EMCStorageConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/ReturnToStoragePool", req, &resp)
	return resp, err
}

//////////////////////////////////////////////////////////////
//             REQUEST Structs used for any                 //
//          masking view deletion on the VMAX3.             //
//////////////////////////////////////////////////////////////

type DeleteMaskingViewReq struct {
	DeleteMaskingViewRequestContent *DeleteMaskingViewReqContent `json:"content"`
}

type DeleteMaskingViewReqContent struct {
	AtType                            string                         `json:"@type"`
	DeleteMaskingViewRequestContentPC *DeleteMaskingViewReqContentPC `json:"ProtocolController"`
}

type DeleteMaskingViewReqContentPC struct {
	AtType                  string `json:"@type"`
	DeviceID                string `json:"DeviceID"`
	CreationClassName       string `json:"CreationClassName"`
	SystemName              string `json:"SystemName"`
	SystemCreationClassName string `json:"SystemCreationClassName"`
}

////////////////////////////////////////////////////////////
//           RESPONSE Struct used for any                 //
//        masking view deletion on the VMAX3.             //
////////////////////////////////////////////////////////////

type DeleteMaskingViewResp struct {
	Entries []struct {
		Content struct {
			AtType       string `json:"@type"`
			I_parameters struct {
				I_Job struct {
					AtType        string `json:"@type"`
					E0_InstanceID string `json:"e0$InstanceID"`
					Xmlns_e0      string `json:"xmlns$e0"`
				} `json:"i$Job"`
			} `json:"i$parameters"`
			I_returnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

/////////////////////////////////////////////////////////////////
//               DELETE a Masking View                         //
/////////////////////////////////////////////////////////////////

func (smis *SMIS) PostDeleteMaskingView(req *DeleteMaskingViewReq, sid string) (resp *DeleteMaskingViewResp, err error) {
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_ControllerConfigurationService/CreationClassName::Symm_ControllerConfigurationService,Name::EMCControllerConfigurationService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/DeleteMaskingView", req, &resp)
	return resp, err
}

/////////////////////////////////////////////////////////
//               REQUEST Structs used for              //
//        getting Ports logged in, on the VMAX3.       //
/////////////////////////////////////////////////////////

type PostPortLoggedInReq struct {
	PostPortLoggedInRequestContent *PostPortLoggedInReqContent `json:"content"`
}

type PostPortLoggedInReqContent struct {
	PostPortLoggedInRequestHardwareID *PostPortLoggedInReqHardwareID `json:"HardwareID"`
	AtType                            string                         `json:"@type"`
}

type PostPortLoggedInReqHardwareID struct {
	AtType     string `json:"@type"`
	InstanceID string `json: "InstanceID"`
}

/////////////////////////////////////////////////////////
//               RESPONSE Structs used for             //
//        getting Ports logged in, on the VMAX3.       //
/////////////////////////////////////////////////////////

type PostPortLoginResp struct {
	Entries []struct {
		Content struct {
			_type        string `json:"@type"`
			I_parameters struct {
				I_TargetEndpoints []map[string]string `json:"i$TargetEndpoints"`
			} `json:"i$parameters"`
			I_returnValue int    `json:"i$returnValue"`
			Xmlns_i       string `json:"xmlns$i"`
		} `json:"content"`
		Content_type string `json:"content-type"`
		Links        []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		Updated string `json:"updated"`
	} `json:"entries"`
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Updated  string `json:"updated"`
	Xmlns_gd string `json:"xmlns$gd"`
}

type PortValues struct {
	WWN        string
	PortNumber string
	Director   string
}

///////////////////////////////////////////////////////////////
//           Get Director Ports                              //
///////////////////////////////////////////////////////////////
func (smis *SMIS) GetTargetEndpoints(systemInstanceName gowbem.CIMInstanceName) ([]gowbem.CIMInstanceName, error) {
	processorSystems, err := smis.EnumerateInstanceNames("Symm_StorageProcessorSystem")
	if err != nil {
		return nil, err
	}

	sysName, _ := smis.GetKeyFromInstanceName(systemInstanceName, "Name")

	results := make([]gowbem.CIMInstanceName, 0)
	for _, service := range processorSystems {
		name, _ := smis.GetKeyFromInstanceName(service, "Name")
		if strings.HasPrefix(name.(string), sysName.(string)) {
			results = append(results, service)
		}
	}
	return results, nil
}

///////////////////////////////////////////////////////////////
//           Getting Ports Logged In                         //
///////////////////////////////////////////////////////////////
func (smis *SMIS) PostPortLogins(req *PostPortLoggedInReq, sid string) (portvalues1 []PortValues, err error) {

	var resp *PostPortLoginResp
	var wwn string = req.PostPortLoggedInRequestContent.PostPortLoggedInRequestHardwareID.InstanceID
	wwn = "W-+-" + wwn
	req.PostPortLoggedInRequestContent.PostPortLoggedInRequestHardwareID.InstanceID = wwn
	err = smis.query("POST", "/ecom/edaa/root/emc/instances/Symm_StorageHardwareIDManagementService/CreationClassName::Symm_StorageHardwareIDManagementService,Name::EMCStorageHardwareIDManagementService,SystemCreationClassName::Symm_StorageSystem,SystemName::"+sid+"/action/EMCGetTargetEndpoints", req, &resp)
	var portValues []PortValues
	var length = len(resp.Entries[0].Content.I_parameters.I_TargetEndpoints)
	for i := 0; i < length; i++ {
		var m map[string]string
		var name string
		m = resp.Entries[0].Content.I_parameters.I_TargetEndpoints[i]
		name = "e" + strconv.Itoa(i) + "$SystemName"
		var eSystemName string = m[name]
		eSystemNameSplit := strings.Split(eSystemName, "-+-")
		PortAndDirector := strings.Split(eSystemNameSplit[2], "-")
		portNumber := PortAndDirector[0]
		director := PortAndDirector[1]
		PV := PortValues{
			WWN:        wwn,
			PortNumber: portNumber,
			Director:   director,
		}
		portValues = append(portValues, PV)
	}
	return portValues, err
}
