//nolint
package asset

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/irisnet/irishub/app/v1/bank"
	sdk "github.com/irisnet/irishub/types"
)

// fee factor formula: (ln(len({name}))/ln{base})^{exp}
const (
	FeeFactorBase = 3
	FeeFactorExp  = 4
)

// GatewayCreateFeeHandler performs fee handling for creating a gateway
func GatewayCreateFeeHandler(ctx sdk.Context, k Keeper, owner sdk.AccAddress, moniker string) sdk.Error {
	// get the required creation fee
	fee := getGatewayCreateFee(ctx, k, moniker)

	return feeHandler(ctx, k, owner, fee)
}

// TokenIssueFeeHandler performs fee handling for issuing token
func TokenIssueFeeHandler(ctx sdk.Context, k Keeper, owner sdk.AccAddress, symbol string) sdk.Error {
	// get the required issuance fee
	fee := getTokenIssueFee(ctx, k, symbol)

	return feeHandler(ctx, k, owner, fee)
}

// TokenMintFeeHandler performs fee handling for minting token
func TokenMintFeeHandler(ctx sdk.Context, k Keeper, owner sdk.AccAddress, symbol string) sdk.Error {
	// get the required minting fee
	fee := getTokenMintFee(ctx, k, symbol)

	return feeHandler(ctx, k, owner, fee)
}

// GatewayTokenIssueFeeHandler performs fee handling for issuing gateway token
func GatewayTokenIssueFeeHandler(ctx sdk.Context, k Keeper, owner sdk.AccAddress, symbol string) sdk.Error {
	// get the required issuance fee
	fee := getGatewayTokenIssueFee(ctx, k, symbol)

	return feeHandler(ctx, k, owner, fee)
}

// GatewayTokenMintFeeHandler performs fee handling for minting gateway token
func GatewayTokenMintFeeHandler(ctx sdk.Context, k Keeper, owner sdk.AccAddress, symbol string) sdk.Error {
	// get the required minting fee
	fee := getGatewayTokenMintFee(ctx, k, symbol)

	return feeHandler(ctx, k, owner, fee)
}

// feeHandler handles the fee of gateway or asset
func feeHandler(ctx sdk.Context, k Keeper, feeAcc sdk.AccAddress, fee sdk.Coin) sdk.Error {
	params := k.GetParamSet(ctx)
	assetTaxRate := params.AssetTaxRate

	// compute community tax and burned coin
	communityTaxCoin := sdk.NewCoin(fee.Denom, sdk.NewDecFromInt(fee.Amount).Mul(assetTaxRate).TruncateInt())
	burnedCoin := fee.Minus(communityTaxCoin)

	// send community tax
	if _, err := k.bk.SendCoins(ctx, feeAcc, bank.CommunityTaxCoinsAccAddr, sdk.Coins{communityTaxCoin}); err != nil {
		return err
	}

	// burn burnedCoin
	if _, err := k.bk.BurnCoins(ctx, feeAcc, sdk.Coins{burnedCoin}); err != nil {
		return err
	}

	return nil
}

// getGatewayCreateFee returns the gateway creation fee
func getGatewayCreateFee(ctx sdk.Context, k Keeper, moniker string) sdk.Coin {
	// get params
	params := k.GetParamSet(ctx)
	gatewayBaseFee := params.CreateGatewayBaseFee

	// compute the fee
	fee := calcFeeByBase(moniker, gatewayBaseFee.Amount)

	return sdk.NewCoin(sdk.NativeTokenMinDenom, convertFeeToInt(fee))
}

// getTokenIssueFee returns the token issurance fee
func getTokenIssueFee(ctx sdk.Context, k Keeper, symbol string) sdk.Coin {
	// get params
	params := k.GetParamSet(ctx)
	issueTokenBaseFee := params.IssueTokenBaseFee

	// compute the fee
	fee := calcFeeByBase(symbol, issueTokenBaseFee.Amount)

	return sdk.NewCoin(sdk.NativeTokenMinDenom, convertFeeToInt(fee))
}

// getTokenMintFee returns the token mint fee
func getTokenMintFee(ctx sdk.Context, k Keeper, symbol string) sdk.Coin {
	// get params
	params := k.GetParamSet(ctx)
	mintTokenFeeRatio := params.MintTokenFeeRatio

	// compute the issurance fee and mint fee
	issueFee := getTokenIssueFee(ctx, k, symbol)
	mintFee := sdk.NewDecFromInt(issueFee.Amount).Mul(mintTokenFeeRatio)

	return sdk.NewCoin(sdk.NativeTokenMinDenom, convertFeeToInt(mintFee))
}

// getGatewayTokenIssueFee returns the gateway token issurance fee
func getGatewayTokenIssueFee(ctx sdk.Context, k Keeper, symbol string) sdk.Coin {
	// get params
	params := k.GetParamSet(ctx)
	gatewayAssetFeeRatio := params.GatewayAssetFeeRatio

	// compute the native token issurance fee and gateway token issurance fee
	nativeTokenIssueFee := getTokenIssueFee(ctx, k, symbol)
	gatewayTokenIssueFee := sdk.NewDecFromInt(nativeTokenIssueFee.Amount).Mul(gatewayAssetFeeRatio)

	return sdk.NewCoin(sdk.NativeTokenMinDenom, convertFeeToInt(gatewayTokenIssueFee))
}

// getGatewayTokenMintFee returns the gateway token mint fee
func getGatewayTokenMintFee(ctx sdk.Context, k Keeper, symbol string) sdk.Coin {
	// get params
	params := k.GetParamSet(ctx)
	gatewayAssetFeeRatio := params.GatewayAssetFeeRatio

	// compute the native token mint fee and gateway token mint fee
	nativeTokenMintFee := getTokenMintFee(ctx, k, symbol)
	gatewayTokenMintFee := sdk.NewDecFromInt(nativeTokenMintFee.Amount).Mul(gatewayAssetFeeRatio)

	return sdk.NewCoin(sdk.NativeTokenMinDenom, convertFeeToInt(gatewayTokenMintFee))
}

// calcFeeByBase computes the actual fee according to the given base fee
func calcFeeByBase(name string, baseFee sdk.Int) sdk.Dec {
	feeFactor := calcFeeFactor(name)
	actualFee := sdk.NewDecFromInt(baseFee).Quo(feeFactor)

	return actualFee
}

// calcFeeFactor computes the fee factor of the given name(common for gateway and asset)
// Note: make sure that the name size is examined before invoking the function
func calcFeeFactor(name string) sdk.Dec {
	nameLen := len(name)
	if nameLen == 0 {
		panic("the length of name must be greater than 0")
	}

	denominator := math.Log(FeeFactorBase)
	numerator := math.Log(float64(nameLen))

	feeFactor := math.Pow(numerator/denominator, FeeFactorExp)
	feeFactorDec, err := sdk.NewDecFromStr(strconv.FormatFloat(feeFactor, 'f', 2, 64))
	if err != nil {
		panic("invalid string")
	}

	return feeFactorDec
}

// convertFeeToInt converts the given fee to Int.
// if greater than 1, rounds it; returns 1 otherwise
func convertFeeToInt(fee sdk.Dec) sdk.Int {
	feeNativeToken := fee.Quo(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 18)))

	if feeNativeToken.GT(sdk.NewDec(1)) {
		return feeNativeToken.TruncateInt().Mul(sdk.NewIntWithDecimal(1, 18))
	} else {
		return sdk.NewInt(1).Mul(sdk.NewIntWithDecimal(1, 18))
	}
}

// GatewayFeeOutput is for the gateway fee query output
type GatewayFeeOutput struct {
	Exist bool     `json:"exist"` // indicate if the gateway has existed
	Fee   sdk.Coin `json:"fee"`   // creation fee
}

// String implements stringer
func (gfo GatewayFeeOutput) String() string {
	var out strings.Builder
	if gfo.Exist {
		out.WriteString("The gateway moniker has existed\n")
	}

	out.WriteString(fmt.Sprintf("Fee: %s", gfo.Fee.String()))

	return out.String()
}

// TokenFeesOutput is for the token fees query output
type TokenFeesOutput struct {
	Exist    bool     `exist`            // indicate if the token has existed
	IssueFee sdk.Coin `json:"issue_fee"` // issue fee
	MintFee  sdk.Coin `json:"mint_fee"`  // mint fee
}

// String implements stringer
func (tfo TokenFeesOutput) String() string {
	var out strings.Builder
	if tfo.Exist {
		out.WriteString("The token id has existed\n")
	}

	out.WriteString(fmt.Sprintf(`Fees:
  IssueFee: %s
  MintFee:  %s`,
		tfo.IssueFee.String(), tfo.MintFee.String()))

	return out.String()
}
