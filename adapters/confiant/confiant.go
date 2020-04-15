// Contributed by Confiant

package confiant

import (
	"fmt"
	"encoding/json"
	"net/http"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/adapters"
)

type ConfiantAdapter struct {
	http           *adapters.HTTPAdapter
	URI            string
}

type confiantBidExtendedConfiant struct {
	UID           string                 `json:"uid"`
}

type confiantBidExtended struct {
	Confiant confiantBidExtendedConfiant `json:"confiant"`
}

func (adapter *ConfiantAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	requests := make([]*adapters.RequestData, 0, 1)
	errors := make([]error, 0, len(request.Imp))
	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")

	jsonRequest, jsonError := json.Marshal(request);
	if jsonError != nil {
		errors = append(errors, jsonError)
	} else {
		var request = &adapters.RequestData {
			Method: "POST",
			Uri: adapter.URI,
			Body: jsonRequest,
			Headers: headers,
		}
		requests = append(requests, request)
	}

	return requests, errors
}

func (adapter *ConfiantAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	switch response.StatusCode {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusBadRequest:
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d.", response.StatusCode),
		}}
	case http.StatusOK:
		var parsedResponse openrtb.BidResponse
		if err := json.Unmarshal(response.Body, &parsedResponse); err != nil {
			return nil, []error{err}
		} else {
			bidsResponse := adapters.NewBidderResponseWithBidsCapacity(len(parsedResponse.SeatBid))
			for _, seatBid := range parsedResponse.SeatBid {
				for i := 0; i < len(seatBid.Bid); i++ {
					bid := seatBid.Bid[i]
					bidsResponse.Bids = append(bidsResponse.Bids, &adapters.TypedBid{Bid: &bid, BidType: openrtb_ext.BidTypeBanner})
				}
			}
			// fmt.Println(string(bidsResponse))
			return bidsResponse, nil
		}
	default:
		return nil, []error{fmt.Errorf("Unexpected status code: %d.", response.StatusCode)}
	}
}

func NewConfiantBidder(client *http.Client, endpoint, platformID string) *ConfiantAdapter {
	adapter := &adapters.HTTPAdapter{Client: client}
	return &ConfiantAdapter{
		http:           adapter,
		URI:            endpoint,
	}
}
