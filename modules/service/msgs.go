package service

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/irisnet/irishub/tools/protoidl"
	"regexp"
)

const (
	// name to idetify transaction types
	MsgType       = "service"
	outputPrivacy = "output_privacy"
	outputCached  = "output_cached"
	description   = "description"
)

var _ sdk.Msg = MsgSvcDef{}

//______________________________________________________________________

// MsgSvcDef - struct for define a service
type MsgSvcDef struct {
	SvcDef
}

func NewMsgSvcDef(name, chainId, description string, tags []string, author sdk.AccAddress, authorDescription, idlContent string) MsgSvcDef {
	return MsgSvcDef{
		SvcDef{
			Name:              name,
			ChainId:           chainId,
			Description:       description,
			Tags:              tags,
			Author:            author,
			AuthorDescription: authorDescription,
			IDLContent:        idlContent,
		},
	}
}

func (msg MsgSvcDef) Route() string { return MsgType }
func (msg MsgSvcDef) Type() string  { return "service definition" }

func (msg MsgSvcDef) GetSignBytes() []byte {
	if len(msg.Tags) == 0 {
		msg.Tags = nil
	}
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgSvcDef) ValidateBasic() sdk.Error {
	if len(msg.ChainId) == 0 {
		return ErrInvalidChainId(DefaultCodespace)
	}
	if !validServiceName(msg.Name) {
		return ErrInvalidServiceName(DefaultCodespace, msg.Name)
	}
	if len(msg.Author) == 0 {
		return ErrInvalidAuthor(DefaultCodespace)
	}
	if len(msg.IDLContent) == 0 {
		return ErrInvalidIDL(DefaultCodespace, "content is empty")
	}
	methods, err := protoidl.GetMethods(msg.IDLContent)
	if err != nil {
		return ErrInvalidIDL(DefaultCodespace, err.Error())
	}
	if valid, err := validateMethods(methods); !valid {
		return err
	}

	return nil
}

func (msg MsgSvcDef) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Author}
}

func validateMethods(methods []protoidl.Method) (bool, sdk.Error) {
	for _, method := range methods {
		if len(method.Name) == 0 {
			return false, ErrInvalidMethodName(DefaultCodespace)
		}
		if _, ok := method.Attributes[outputPrivacy]; ok {
			_, err := OutputPrivacyEnumFromString(method.Attributes[outputPrivacy])
			if err != nil {
				return false, ErrInvalidOutputPrivacyEnum(DefaultCodespace, method.Attributes[outputPrivacy])
			}
		}
		if _, ok := method.Attributes[outputCached]; ok {
			_, err := OutputCachedEnumFromString(method.Attributes[outputCached])
			if err != nil {
				return false, ErrInvalidOutputCachedEnum(DefaultCodespace, method.Attributes[outputCached])
			}
		}
	}
	return true, nil
}

func methodToMethodProperty(index int, method protoidl.Method) (methodProperty MethodProperty, err sdk.Error) {
	// set default value
	opp := NoPrivacy
	opc := NoCached

	var err1 error
	if _, ok := method.Attributes[outputPrivacy]; ok {
		opp, err1 = OutputPrivacyEnumFromString(method.Attributes[outputPrivacy])
		if err1 != nil {
			return methodProperty, ErrInvalidOutputPrivacyEnum(DefaultCodespace, method.Attributes[outputPrivacy])
		}
	}
	if _, ok := method.Attributes[outputCached]; ok {
		opc, err1 = OutputCachedEnumFromString(method.Attributes[outputCached])
		if err != nil {
			return methodProperty, ErrInvalidOutputCachedEnum(DefaultCodespace, method.Attributes[outputCached])
		}
	}
	methodProperty = MethodProperty{
		ID:            index,
		Name:          method.Name,
		Description:   method.Attributes[description],
		OutputPrivacy: opp,
		OutputCached:  opc,
	}
	return
}

//______________________________________________________________________

// MsgSvcBinding - struct for bind a service
type MsgSvcBind struct {
	DefName     string         `json:"def_name"`
	DefChainID  string         `json:"def_chain_id"`
	BindChainID string         `json:"bind_chain_id"`
	Provider    sdk.AccAddress `json:"provider"`
	BindingType BindingType    `json:"binding_type"`
	Deposit     sdk.Coins      `json:"deposit"`
	Prices      []sdk.Coin     `json:"price"`
	Level       Level          `json:"level"`
}

func NewMsgSvcBind(defChainID, defName, bindChainID string, provider sdk.AccAddress, bindingType BindingType, deposit sdk.Coins, prices []sdk.Coin, level Level) MsgSvcBind {
	return MsgSvcBind{
		DefChainID:  defChainID,
		DefName:     defName,
		BindChainID: bindChainID,
		Provider:    provider,
		BindingType: bindingType,
		Deposit:     deposit,
		Prices:      prices,
		Level:       level,
	}
}

func (msg MsgSvcBind) Route() string { return MsgType }
func (msg MsgSvcBind) Type() string  { return "service binding" }

func (msg MsgSvcBind) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgSvcBind) ValidateBasic() sdk.Error {
	if len(msg.DefChainID) == 0 {
		return ErrInvalidDefChainId(DefaultCodespace)
	}
	if len(msg.BindChainID) == 0 {
		return ErrInvalidChainId(DefaultCodespace)
	}
	if !validServiceName(msg.DefName) {
		return ErrInvalidServiceName(DefaultCodespace, msg.DefName)
	}
	if !validBindingType(msg.BindingType) {
		return ErrInvalidBindingType(DefaultCodespace, msg.BindingType)
	}
	if len(msg.Provider) == 0 {
		sdk.ErrInvalidAddress(msg.Provider.String())
	}
	if !msg.Deposit.IsValid() {
		return sdk.ErrInvalidCoins(msg.Deposit.String())
	}
	if !msg.Deposit.IsNotNegative() {
		return sdk.ErrInvalidCoins(msg.Deposit.String())
	}
	for _, price := range msg.Prices {
		if !price.IsNotNegative() {
			return sdk.ErrInvalidCoins(price.String())
		}
	}
	if !validLevel(msg.Level) {
		return ErrInvalidLevel(DefaultCodespace, msg.Level)
	}
	return nil
}

func (msg MsgSvcBind) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Provider}
}

//______________________________________________________________________

// MsgSvcBindingUpdate - struct for update a service binding
type MsgSvcBindingUpdate struct {
	DefName     string         `json:"def_name"`
	DefChainID  string         `json:"def_chain_id"`
	BindChainID string         `json:"bind_chain_id"`
	Provider    sdk.AccAddress `json:"provider"`
	BindingType BindingType    `json:"binding_type"`
	Deposit     sdk.Coins      `json:"deposit"`
	Prices      []sdk.Coin     `json:"price"`
	Level       Level          `json:"level"`
}

func NewMsgSvcBindingUpdate(defChainID, defName, bindChainID string, provider sdk.AccAddress, bindingType BindingType, deposit sdk.Coins, prices []sdk.Coin, level Level) MsgSvcBindingUpdate {
	return MsgSvcBindingUpdate{
		DefChainID:  defChainID,
		DefName:     defName,
		BindChainID: bindChainID,
		Provider:    provider,
		BindingType: bindingType,
		Deposit:     deposit,
		Prices:      prices,
		Level:       level,
	}
}
func (msg MsgSvcBindingUpdate) Route() string { return MsgType }
func (msg MsgSvcBindingUpdate) Type() string  { return "service binding update" }

func (msg MsgSvcBindingUpdate) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgSvcBindingUpdate) ValidateBasic() sdk.Error {
	if len(msg.DefChainID) == 0 {
		return ErrInvalidDefChainId(DefaultCodespace)
	}
	if len(msg.BindChainID) == 0 {
		return ErrInvalidChainId(DefaultCodespace)
	}
	if !validServiceName(msg.DefName) {
		return ErrInvalidServiceName(DefaultCodespace, msg.DefName)
	}
	if len(msg.Provider) == 0 {
		sdk.ErrInvalidAddress(msg.Provider.String())
	}
	if msg.BindingType != 0x00 && !validBindingType(msg.BindingType) {
		return ErrInvalidBindingType(DefaultCodespace, msg.BindingType)
	}
	if !msg.Deposit.IsNotNegative() {
		return sdk.ErrInvalidCoins(msg.Deposit.String())
	}
	for _, price := range msg.Prices {
		if !price.IsNotNegative() {
			return sdk.ErrInvalidCoins(price.String())
		}
	}
	if !validUpdateLevel(msg.Level) {
		return ErrInvalidLevel(DefaultCodespace, msg.Level)
	}
	return nil
}

func (msg MsgSvcBindingUpdate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Provider}
}

//______________________________________________________________________

// MsgSvcDisable - struct for disable a service binding
type MsgSvcDisable struct {
	DefName     string         `json:"def_name"`
	DefChainID  string         `json:"def_chain_id"`
	BindChainID string         `json:"bind_chain_id"`
	Provider    sdk.AccAddress `json:"provider"`
}

func NewMsgSvcDisable(defChainID, defName, bindChainID string, provider sdk.AccAddress) MsgSvcDisable {
	return MsgSvcDisable{
		DefChainID:  defChainID,
		DefName:     defName,
		BindChainID: bindChainID,
		Provider:    provider,
	}
}

func (msg MsgSvcDisable) Route() string { return MsgType }
func (msg MsgSvcDisable) Type() string  { return "service disable" }

func (msg MsgSvcDisable) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgSvcDisable) ValidateBasic() sdk.Error {
	if len(msg.DefChainID) == 0 {
		return ErrInvalidDefChainId(DefaultCodespace)
	}
	if len(msg.BindChainID) == 0 {
		return ErrInvalidChainId(DefaultCodespace)
	}
	if !validServiceName(msg.DefName) {
		return ErrInvalidServiceName(DefaultCodespace, msg.DefName)
	}
	if len(msg.Provider) == 0 {
		sdk.ErrInvalidAddress(msg.Provider.String())
	}
	return nil
}

func (msg MsgSvcDisable) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Provider}
}

//______________________________________________________________________

// MsgSvcEnable - struct for enable a service binding
type MsgSvcEnable struct {
	DefName     string         `json:"def_name"`
	DefChainID  string         `json:"def_chain_id"`
	BindChainID string         `json:"bind_chain_id"`
	Provider    sdk.AccAddress `json:"provider"`
	Deposit     sdk.Coins      `json:"deposit"`
}

func NewMsgSvcEnable(defChainID, defName, bindChainID string, provider sdk.AccAddress, deposit sdk.Coins) MsgSvcEnable {
	return MsgSvcEnable{
		DefChainID:  defChainID,
		DefName:     defName,
		BindChainID: bindChainID,
		Provider:    provider,
		Deposit:     deposit,
	}
}

func (msg MsgSvcEnable) Route() string { return MsgType }
func (msg MsgSvcEnable) Type() string  { return "service enable" }

func (msg MsgSvcEnable) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgSvcEnable) ValidateBasic() sdk.Error {
	if len(msg.DefChainID) == 0 {
		return ErrInvalidDefChainId(DefaultCodespace)
	}
	if len(msg.BindChainID) == 0 {
		return ErrInvalidChainId(DefaultCodespace)
	}
	if !validServiceName(msg.DefName) {
		return ErrInvalidServiceName(DefaultCodespace, msg.DefName)
	}
	if !msg.Deposit.IsNotNegative() {
		return sdk.ErrInvalidCoins(msg.Deposit.String())
	}
	if len(msg.Provider) == 0 {
		sdk.ErrInvalidAddress(msg.Provider.String())
	}
	return nil
}

func (msg MsgSvcEnable) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Provider}
}

//______________________________________________________________________

// MsgSvcRefundDeposit - struct for refund deposit from a service binding
type MsgSvcRefundDeposit struct {
	DefName     string         `json:"def_name"`
	DefChainID  string         `json:"def_chain_id"`
	BindChainID string         `json:"bind_chain_id"`
	Provider    sdk.AccAddress `json:"provider"`
}

func NewMsgSvcRefundDeposit(defChainID, defName, bindChainID string, provider sdk.AccAddress) MsgSvcRefundDeposit {
	return MsgSvcRefundDeposit{
		DefChainID:  defChainID,
		DefName:     defName,
		BindChainID: bindChainID,
		Provider:    provider,
	}
}

func (msg MsgSvcRefundDeposit) Route() string { return MsgType }
func (msg MsgSvcRefundDeposit) Type() string  { return "service refund deposit" }

func (msg MsgSvcRefundDeposit) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgSvcRefundDeposit) ValidateBasic() sdk.Error {
	if len(msg.DefChainID) == 0 {
		return ErrInvalidDefChainId(DefaultCodespace)
	}
	if len(msg.BindChainID) == 0 {
		return ErrInvalidChainId(DefaultCodespace)
	}
	if !validServiceName(msg.DefName) {
		return ErrInvalidServiceName(DefaultCodespace, msg.DefName)
	}
	if len(msg.Provider) == 0 {
		sdk.ErrInvalidAddress(msg.Provider.String())
	}
	return nil
}

func (msg MsgSvcRefundDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Provider}
}

func validServiceName(name string) bool {
	if len(name) == 0 || len(name) > 128 {
		return false
	}

	// Must contain alphanumeric characters, _ and - only
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return !reg.Match([]byte(name))
}
