package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/StackExchange/wmi"

	rpc "github.com/disaipe/dev01-rpc-base"
)

type GetComputerStateRequest struct {
	rpc.Response

	Id    int
	Host  string
	Hosts []GetComputerStateRequest
}

type GetComputerStateResponse struct {
	rpc.ResultResponse

	Id       int
	Status   bool
	UserName string
	Hosts    []GetComputerStateResponse
}

type Win32_ComputerSystem struct {
	Name     string
	UserName string
}

func GetComputerState(computerStateRequest GetComputerStateRequest) GetComputerStateResponse {
	var dst []Win32_ComputerSystem

	query := wmi.CreateQuery(&dst, "")

	err := wmi.Query(query, &dst, computerStateRequest.Host, "root\\CIMV2")

	computerStateResponse := GetComputerStateResponse{
		Id:     computerStateRequest.Id,
		Status: true,
	}

	if err != nil {
		rpc.Logger.Error().Msgf("Failed to query host %s: %v", computerStateRequest.Host, err)
		computerStateResponse.Status = false
	} else {
		rpc.Logger.Info().Msgf("Query host %s successful", computerStateRequest.Host)
		computerStateResponse.UserName = dst[0].UserName
	}

	return computerStateResponse
}

func GetBatchComputersStateJob(rpcServer *rpc.Rpc, computerStateRequest GetComputerStateRequest, appAuth string) rpc.Job {
	return rpc.Job{
		Name: fmt.Sprintf("chunk-%d", computerStateRequest.Id),
		Action: func() error {
			computerStateResponse := GetComputerStateResponse{
				Status: true,
			}

			total := len(computerStateRequest.Hosts)

			for key, host := range computerStateRequest.Hosts {
				if host.Id != 0 && host.Host != "" {
					hostResult := GetComputerState(host)
					computerStateResponse.Hosts = append(computerStateResponse.Hosts, hostResult)

					rpc.Logger.Info().Msgf("%d/%d end", key, total)
				}
			}

			rpcServer.SendResult(computerStateResponse, appAuth)

			return nil
		},
	}
}

var rpcAction = rpc.ActionFunction(func(rpcServer *rpc.Rpc, body io.ReadCloser, appAuth string) (rpc.Response, error) {
	var computerStateRequest GetComputerStateRequest

	err := json.NewDecoder(body).Decode(&computerStateRequest)

	if err != nil {
		return nil, err
	}

	var resultStatus = true
	var resultMessage string

	if computerStateRequest.Hosts != nil {
		rpcServer.AddJob(GetBatchComputersStateJob(rpcServer, computerStateRequest, appAuth))
	} else {
		if computerStateRequest.Id == 0 {
			resultStatus = false
			resultMessage = "Id is required"
		} else if computerStateRequest.Host == "" {
			resultStatus = false
			resultMessage = "Host is required"
		} else {
			go func() {
				result := GetComputerState(computerStateRequest)

				rpcServer.SendResult(result, appAuth)
			}()
		}
	}

	requestAcceptedResponse := &rpc.ActionResponse{
		Status: resultStatus,
		Data:   resultMessage,
	}

	return requestAcceptedResponse, nil
})

func main() {
	flag.Parse()

	rpc.Config.SetServiceSettings(
		"dev01-pc-monitor",
		"Dev01 Computer state monitor",
		"The part of the Dev01 platform",
	)

	rpc.Config.SetAction("/computer_sync_job", &rpcAction)

	if rpc.Config.Serving() {
		rpcServer := &rpc.Rpc{}
		rpcServer.Run()
	}
}
