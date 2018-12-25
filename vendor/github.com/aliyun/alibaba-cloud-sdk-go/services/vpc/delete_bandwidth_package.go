package vpc

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

func (client *Client) DeleteBandwidthPackage(request *DeleteBandwidthPackageRequest) (response *DeleteBandwidthPackageResponse, err error) {
	response = CreateDeleteBandwidthPackageResponse()
	err = client.DoAction(request, response)
	return
}

func (client *Client) DeleteBandwidthPackageWithChan(request *DeleteBandwidthPackageRequest) (<-chan *DeleteBandwidthPackageResponse, <-chan error) {
	responseChan := make(chan *DeleteBandwidthPackageResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DeleteBandwidthPackage(request)
		responseChan <- response
		errChan <- err
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

func (client *Client) DeleteBandwidthPackageWithCallback(request *DeleteBandwidthPackageRequest, callback func(response *DeleteBandwidthPackageResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DeleteBandwidthPackageResponse
		var err error
		defer close(result)
		response, err = client.DeleteBandwidthPackage(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

type DeleteBandwidthPackageRequest struct {
	*requests.RpcRequest
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	Force                requests.Boolean `position:"Query" name:"Force"`
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	BandwidthPackageId   string           `position:"Query" name:"BandwidthPackageId"`
}

type DeleteBandwidthPackageResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

func CreateDeleteBandwidthPackageRequest() (request *DeleteBandwidthPackageRequest) {
	request = &DeleteBandwidthPackageRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Vpc", "2016-04-28", "DeleteBandwidthPackage", "vpc", "openAPI")
	return
}

func CreateDeleteBandwidthPackageResponse() (response *DeleteBandwidthPackageResponse) {
	response = &DeleteBandwidthPackageResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
