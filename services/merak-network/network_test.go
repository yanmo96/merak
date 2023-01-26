package main

import (
	"fmt"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	"github.com/futurewei-cloud/merak/services/merak-network/activities"
	"github.com/futurewei-cloud/merak/services/merak-network/database"
	"github.com/futurewei-cloud/merak/services/merak-network/utils"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

//func TestNetworkCreate(t *testing.T) {
//
//}

func TestNodeRegister(t *testing.T) {
	log.Println("start TestNodeRegister")
	tests := []struct {
		giveNodeBody    []*common_pb.InternalComputeInfo
		giveNetConfigId string
		server          *httptest.Server
		expRes          string
		expErr          error
		pass            bool
	}{
		{
			giveNodeBody: []*common_pb.InternalComputeInfo{
				{
					OperationType: common_pb.OperationType_CREATE,
					Id:            "YM_node5_id",
					Name:          "YM_node5",
					DatapathIp:    "192.168.10.15",
					Mac:           "36:db:23:8c:4a:c5",
					Veth:          "eth1",
					ContainerIp:   "10.1.0.5",
				},
				{
					OperationType: common_pb.OperationType_CREATE,
					Id:            "YM_node6_id",
					Name:          "YM_node6",
					DatapathIp:    "192.168.10.16",
					Mac:           "36:db:23:8c:4a:c6",
					Veth:          "eth1",
					ContainerIp:   "10.1.0.6",
				},
			},
			giveNetConfigId: "1234",
			server: httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusCreated)
				writer.Write([]byte(`[{"ncm_uri":"null/bulk","node_id":"YM_node5_id","node_name":"YM_node5","local_ip":"10.1.0.5","mac_address":"36:db:23:8c:4a:c5","veth":"eth1","server_port":50001,"host_dvr_mac":null,"ncm_id":null,"data-path-ip":"192.168.10.15"},{"ncm_uri":null,"node_id":"YM_node6_id","node_name":"YM_node6","local_ip":"10.1.0.6","mac_address":"36:db:23:8c:4a:c6","veth":"eth1","server_port":50001,"host_dvr_mac":null,"ncm_id":null,"data-path-ip":"192.168.10.16"}]`))
			})),
			expRes: "RegisterNode done",
			expErr: nil,
			pass:   true,
		},
	}

	utils.REDISADDR = "localhost:55000"
	// Connect to storage
	if err := database.ConnectDatabase(); err != nil {
		fmt.Printf("Cannot connect to Redis db!, error: '%s'\n", err)
	}

	for _, tt := range tests {
		t.Run("register node", func(t *testing.T) {
			defer tt.server.Close()
			utils.ALCORURL = tt.server.URL

			// Do the Register Node activity
			returnString, err := activities.RegisterNode(
				tt.giveNodeBody,
				tt.giveNetConfigId,
			)

			// check response parse
			if !reflect.DeepEqual(tt.expRes, returnString) {
				t.Error("Expected", tt.expRes, "got", returnString)
			}
			if err != nil {
				assert.Equal(t, tt.expErr.Error(), err.Error())
			}
		})
	}
	log.Println("finish TestNodeRegister")
}
