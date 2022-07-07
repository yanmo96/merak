package activities

import (
	"context"
	"encoding/json"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-network/database"
	"github.com/futurewei-cloud/merak/services/merak-network/entities"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
	"github.com/futurewei-cloud/merak/services/merak-network/utils"
	"log"
	"sync"
)

var (
	returnNetworkMessage = pb.ReturnNetworkMessage{
		ReturnCode:       pb.ReturnCode_OK,
		ReturnMessage:    "returnNetworkMessage Finished",
		Vpcs:             nil,
		SecurityGroupIds: nil,
	}
)

func doVPC(vpc *pb.InternalVpcInfo) (vpcId string) {
	log.Println("doVPC")
	vpcBody := entities.VpcStruct{Network: entities.VpcBody{
		AdminStateUp:        true,
		RevisionNumber:      0,
		Cidr:                vpc.VpcCidr,
		ByDefault:           true,
		Description:         "vpc",
		DnsDomain:           "domain",
		IsDefault:           true,
		Mtu:                 1400,
		Name:                "YM_sample_vpc",
		PortSecurityEnabled: true,
		ProjectId:           vpc.ProjectId,
	}}
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30001/project/123456789/vpcs", "POST", vpcBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.VpcReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	database.Set(utils.VPC+returnJson.Network.ID, returnJson.Network)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doVPC done")
	return returnJson.Network.ID
}
func doSubnet(subnet *pb.InternalSubnetInfo, vpcId string) (subnetId string) {
	log.Println("doSubnet")
	subnetBody := entities.SubnetStruct{Subnet: entities.SubnetBody{
		Cider:     subnet.SubnetCidr,
		Name:      "YM_sample_subnet",
		IpVersion: 4,
		NetworkId: vpcId,
	}}
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30002/project/123456789/subnets", "POST", subnetBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.SubnetReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	database.Set(utils.SUBNET+returnJson.Subnet.ID, returnJson.Subnet)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doVPC done")
	return returnJson.Subnet.ID
}
func doRouter(vpcId string) (routerId string) {
	log.Println("doRouter")
	routerBody := entities.RouterStruct{Router: entities.RouterBody{
		AdminStateUp: true,
		Description:  "router description",
		Distributed:  true,
		ExternalGatewayInfo: entities.RouterExternalGatewayInfo{
			EnableSnat:       true,
			ExternalFixedIps: nil,
			NetworkId:        vpcId,
		},
		FlavorId:       "",
		GatewayPorts:   nil,
		Ha:             true,
		Name:           "YM_simple_router",
		ProjectId:      "123456789",
		RevisionNumber: 0,
		Status:         "BUILD",
		TenantId:       "123456789",
	}}
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30003/project/123456789/routers", "POST", routerBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.RouterReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	database.Set(utils.Router+returnJson.Router.ID, returnJson.Router)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doRouter done")
	return returnJson.Router.ID
}
func doAttachRouter(routerId string, subnetId string) error {
	log.Println("doAttachRouter")
	attachRouterBody := entities.AttachRouterStruct{SubnetId: subnetId}
	url := "http://" + utils.ALCORURL + ":30003/project/123456789/routers/" + routerId + "/add_router_interface"
	returnMessage, returnErr := http.RequestCall(url, "PUT", attachRouterBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.AttachRouterReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doAttachRouter done")
	return nil
}
func doSg(sg *pb.InternalSecurityGroupInfo, sgID string) string {
	log.Println("doSg")
	sgBody := entities.SgStruct{Sg: entities.SgBody{
		Id:                 sgID,
		Description:        "sg Description",
		Name:               "YM_sample_sg",
		ProjectId:          sg.ProjectId,
		SecurityGroupRules: nil,
		TenantId:           sg.TenantId,
	}}
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30008/project/123456789/security-groups", "POST", sgBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.SgReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	database.Set(utils.SECURITYGROUP+returnJson.SecurityGroup.ID, returnJson.SecurityGroup)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doSg done")
	return returnJson.SecurityGroup.ID
}

func VnetCreate(ctx context.Context, netConfigId string, network *pb.InternalNetworkInfo, wg *sync.WaitGroup, returnMessage chan *pb.ReturnNetworkMessage) (string, error) {
	log.Println("VnetCreate")
	//defer wg.Done()
	// TODO may want to separate bellow sections to different function, and use `go` and `wg` to improve overall speed
	// TODO when do concurrent, need to keep in mind on how to control the number of concurrency
	// Doing vpc and subnet

	var vpcId string
	var vpcIds []string
	subnetCiderIdMap := make(map[string]string)
	for _, vpc := range network.Vpcs {
		//for i := 0; i < int(network.NumberOfVpcs); i++ {
		vpcId = doVPC(vpc)
		vpcIds = append(vpcIds, vpcId)
		var returnInfo []*pb.InternalVpcInfo

		var subnetInfo []*pb.InternalSubnetInfo
		for _, subnet := range vpc.Subnets {
			//for j := 0; j < int(network.NumberOfSubnetPerVpc); j++ {
			//subnetId := utils.GenUUID()
			subnetId := doSubnet(subnet, vpcId)
			subnetCiderIdMap[subnet.SubnetCidr] = subnetId
			log.Printf("subnetCiderIdMap %s", subnetCiderIdMap)
			currentSubnet := pb.InternalSubnetInfo{
				SubnetId:   subnetId,
				SubnetCidr: subnet.SubnetCidr,
				SubnetGw:   subnet.SubnetGw,
				NumberVms:  subnet.NumberVms,
			}
			subnetInfo = append(subnetInfo, &currentSubnet)
		}
		currentVPC := pb.InternalVpcInfo{
			VpcId:     vpcId,
			TenantId:  vpc.TenantId,
			ProjectId: vpc.ProjectId,
			Subnets:   subnetInfo,
		}
		//returnInfo = append(returnInfo, &currentVPC)
		returnNetworkMessage.Vpcs = append(returnNetworkMessage.Vpcs, &currentVPC)
		//returnNetworkMessage.Vpcs = append(returnNetworkMessage.Vpcs, returnInfo)
		log.Printf("VnetCreate End %s", returnInfo)
	}

	//doing security group
	for i := 0; i < int(network.NumberOfSecurityGroups); i++ {
		sgId := utils.GenUUID()
		go doSg(network.SecurityGroups[i], sgId)
		returnNetworkMessage.SecurityGroupIds = append(returnNetworkMessage.SecurityGroupIds, sgId)
		log.Printf("sgId: %s", sgId)
	}

	//doing router: create and attach subnet
	for _, router := range network.Routers {
		routerId := doRouter(vpcId)
		for _, subnet := range router.Subnets {
			doAttachRouter(routerId, subnetCiderIdMap[subnet])
		}
	}
	database.Set(utils.NETCONFIG+netConfigId, &returnNetworkMessage)
	returnMessage <- &returnNetworkMessage
	log.Printf("&returnNetworkMessage %s", &returnNetworkMessage)
	defer wg.Done()
	return "VnetCreate", nil
}
