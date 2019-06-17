package asset

import (
	"fmt"

	"github.com/irisnet/irishub/app/v1/params"
	"github.com/irisnet/irishub/codec"
	sdk "github.com/irisnet/irishub/types"
)

var _ params.ParamSet = (*Params)(nil)

const (
	DefaultParamSpace = "asset"
)

// parameter keys
var (
	KeyTokenTaxRate         = []byte("TokenTaxRate")
	KeyIssueFTBaseFee       = []byte("IssueFTBaseFee")
	KeyMintFTFeeRatio       = []byte("MintFTFeeRatio")
	KeyCreateGatewayBaseFee = []byte("CreateGatewayBaseFee")
	KeyGatewayAssetFeeRatio = []byte("GatewayAssetFeeRatio")
)

// ParamTable for asset module
func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable().RegisterParamSet(&Params{})
}

// asset params
type Params struct {
	TokenTaxRate         sdk.Dec  `json:"asset_tax_rate"`          // e.g., 40%
	IssueFTBaseFee       sdk.Coin `json:"issue_ft_base_fee"`       // e.g., 300000*10^18iris-atto
	MintFTFeeRatio       sdk.Dec  `json:"mint_ft_fee_ratio"`       // e.g., 10%
	CreateGatewayBaseFee sdk.Coin `json:"create_gateway_base_fee"` // e.g., 600000*10^18iris-atto
	GatewayAssetFeeRatio sdk.Dec  `json:"gateway_asset_fee_ratio"` // e.g., 10%
} // issuance fee = IssueFTBaseFee / (ln(len(symbol))/ln3)^4

func (p Params) String() string {
	return fmt.Sprintf(`Token Params:
  Token Tax Rate:                                           %s
  Base Fee for Issuing Fungible Token:                      %s
  Fee Ratio for Minting (vs Issuing) Fungible Token:        %s
  Base Fee for Creating Gateway:                            %s
  Fee Ratio for Gateway (vs Native) Assets:                 %s`,
		p.TokenTaxRate.String(), p.IssueFTBaseFee.String(), p.MintFTFeeRatio.String(), p.CreateGatewayBaseFee.String(), p.GatewayAssetFeeRatio.String())
}

// Implements params.ParamSet
func (p *Params) GetParamSpace() string {
	return DefaultParamSpace
}

func (p *Params) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{KeyTokenTaxRate, &p.TokenTaxRate},
		{KeyIssueFTBaseFee, &p.IssueFTBaseFee},
		{KeyMintFTFeeRatio, &p.MintFTFeeRatio},
		{KeyCreateGatewayBaseFee, &p.CreateGatewayBaseFee},
		{KeyGatewayAssetFeeRatio, &p.GatewayAssetFeeRatio},
	}
}

func (p *Params) Validate(key string, value string) (interface{}, sdk.Error) {
	switch key {
	case string(KeyTokenTaxRate):
		rate, err := sdk.NewDecFromStr(value)
		if err != nil {
			return nil, params.ErrInvalidString(value)
		}
		if err := validateAssetTaxRate(rate); err != nil {
			return nil, err
		}
		return rate, nil
	case string(KeyIssueFTBaseFee):
		fee, err := sdk.ParseCoin(value)
		if err != nil || fee.Denom != sdk.NativeTokenMinDenom {
			return nil, params.ErrInvalidString(value)
		}
		return fee, nil
	case string(KeyMintFTFeeRatio):
		ratio, err := sdk.NewDecFromStr(value)
		if err != nil {
			return nil, params.ErrInvalidString(value)
		}
		if err := validateMintFTBaseFeeRatio(ratio); err != nil {
			return nil, err
		}
		return ratio, nil
	case string(KeyCreateGatewayBaseFee):
		fee, err := sdk.ParseCoin(value)
		if err != nil || fee.Denom != sdk.NativeTokenMinDenom {
			return nil, params.ErrInvalidString(value)
		}

		return fee, nil
	case string(KeyGatewayAssetFeeRatio):
		ratio, err := sdk.NewDecFromStr(value)
		if err != nil {
			return nil, params.ErrInvalidString(value)
		}
		if err := validateGatewayAssetFeeRatio(ratio); err != nil {
			return nil, err
		}
		return ratio, nil
	default:
		return nil, sdk.NewError(params.DefaultCodespace, params.CodeInvalidKey, fmt.Sprintf("%s is an invalid key", key))
	}
}

func (p *Params) StringFromBytes(cdc *codec.Codec, key string, bytes []byte) (string, error) {
	return "", fmt.Errorf("this method is not implemented")
}

// default asset module params
func DefaultParams() Params {
	return Params{
		TokenTaxRate:         sdk.NewDecWithPrec(4, 1), // 0.4 (40%)
		IssueFTBaseFee:       sdk.NewCoin(sdk.NativeTokenMinDenom, sdk.NewIntWithDecimal(300000, 18)),
		MintFTFeeRatio:       sdk.NewDecWithPrec(1, 1), // 0.1 (10%)
		CreateGatewayBaseFee: sdk.NewCoin(sdk.NativeTokenMinDenom, sdk.NewIntWithDecimal(600000, 18)),
		GatewayAssetFeeRatio: sdk.NewDecWithPrec(1, 1), // 0.1 (10%)
	}
}

// default asset module params for test
func DefaultParamsForTest() Params {
	return Params{
		TokenTaxRate:         sdk.NewDecWithPrec(4, 1), // 0.4 (40%)
		IssueFTBaseFee:       sdk.NewCoin(sdk.NativeTokenMinDenom, sdk.NewIntWithDecimal(300000, 18)),
		MintFTFeeRatio:       sdk.NewDecWithPrec(1, 1), // 0.1 (10%)
		CreateGatewayBaseFee: sdk.NewCoin(sdk.NativeTokenMinDenom, sdk.NewIntWithDecimal(600000, 18)),
		GatewayAssetFeeRatio: sdk.NewDecWithPrec(1, 1), // 0.1 (10%)
	}
}

func validateParams(p Params) error {
	if sdk.NetworkType != sdk.Mainnet {
		return nil
	}

	if err := validateAssetTaxRate(p.TokenTaxRate); err != nil {
		return err
	}
	if err := validateMintFTBaseFeeRatio(p.MintFTFeeRatio); err != nil {
		return err
	}
	if err := validateGatewayAssetFeeRatio(p.GatewayAssetFeeRatio); err != nil {
		return err
	}

	return nil
}

func validateAssetTaxRate(v sdk.Dec) sdk.Error {
	if v.GT(sdk.NewDecWithPrec(1, 0)) || v.LT(sdk.NewDecWithPrec(0, 0)) {
		return sdk.NewError(
			params.DefaultCodespace,
			params.CodeInvalidAssetTaxRate,
			fmt.Sprintf("Token Tax Rate [%s] should be between [0, 1]", v.String()),
		)
	}
	return nil
}

func validateMintFTBaseFeeRatio(v sdk.Dec) sdk.Error {
	if v.GT(sdk.NewDecWithPrec(1, 0)) || v.LT(sdk.NewDecWithPrec(0, 0)) {
		return sdk.NewError(
			params.DefaultCodespace,
			params.CodeInvalidMintFTBaseFeeRatio,
			fmt.Sprintf("Base Fee Ratio for Minting FTs [%s] should be between [0, 1]", v.String()),
		)
	}
	return nil
}

func validateGatewayAssetFeeRatio(v sdk.Dec) sdk.Error {
	if v.GT(sdk.NewDecWithPrec(1, 0)) || v.LT(sdk.NewDecWithPrec(0, 0)) {
		return sdk.NewError(
			params.DefaultCodespace,
			params.CodeInvalidGatewayAssetFeeRatio,
			fmt.Sprintf("Fee Ratio for Gateway Assets [%s] should be between [0, 1]", v.String()),
		)
	}
	return nil
}

// get asset params from the global param store
func (k Keeper) GetParamSet(ctx sdk.Context) Params {
	var p Params
	k.paramSpace.GetParamSet(ctx, &p)
	return p
}

// set asset params from the global param store
func (k Keeper) SetParamSet(ctx sdk.Context, params Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
