package pulse

import (
	"context"
	"math/big"
	"strings"

	"github.com/sajal/mtrparser"
)

type MtrResult struct {
	Result     *mtrparser.MTROutPut
	Err        string
	ErrEnglish string //Human friendly version of Err
}

type MtrRequest struct {
	Target      string
	IPv         string //blank for auto, 4 for IPv4, 6 for IPv6
	AgentFilter []*big.Int
}

func MtrImpl(ctx context.Context, r *MtrRequest) *MtrResult {
	var result MtrResult
	defer translateMtrError(&result)
	//Validate r.Target before sending
	tgt := strings.Trim(r.Target, "\n \r") //Trim whitespace
	if strings.Contains(tgt, " ") {        //Ensure it doesn't contain space
		result.Err = "Invalid hostname"
		return &result
	}
	if strings.HasPrefix(tgt, "-") { //Ensure it doesn't start with -
		result.Err = "Invalid hostname"
		return &result
	}
	out, err := mtrparser.ExecuteMTRContext(ctx, tgt, r.IPv)
	if err != nil {
		result.Err = err.Error()
		return &result
	}
	result.Result = out
	return &result
}
