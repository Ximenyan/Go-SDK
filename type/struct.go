package types

import (
	"cocos-go-sdk/common"
	"encoding/hex"
	"fmt"
	"math/big"

	"strconv"
	"strings"

	"github.com/itchyny/base58-go"
)

func PukBytesFromBase58String(base58Str string) []byte {
	byte_s, _ := base58.BitcoinEncoding.Decode([]byte(base58Str)[5:])
	big_i, _ := new(big.Int).SetString(string(byte_s), 10)
	data := big_i.Bytes()
	puk := data[0 : len(data)-4]
	return puk
}

type ObjectId string
type Object interface {
	GetBytes() []byte
}

func (o ObjectId) GetBytes() []byte {
	num := strings.Split(string(o), `.`)[2]
	i, _ := strconv.ParseUint(num, 10, 64)
	return common.Varint(i)
}

type Amount struct {
	Amount  uint64   `json:"amount"`
	AssetID ObjectId `json:"asset_id"`
}

func (a Amount) GetBytes() []byte {
	byte_s := append(common.VarUint(a.Amount, 64), a.AssetID.GetBytes()...)
	return byte_s
}

type Memo struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Nonce   uint64 `json:"nonce"`
	Message string `json:"message"`
}

func (o Memo) GetBytes() []byte {
	from := PukBytesFromBase58String(o.From)
	to := PukBytesFromBase58String(o.To)
	nonce := common.VarUint(o.Nonce, 64)
	msg, _ := hex.DecodeString(o.Message)
	msg = append([]byte{byte(len(msg))}, msg...)
	byte_s := append([]byte{0x01},
		append(from,
			append(to,
				append(nonce, msg...)...)...)...,
	)
	return byte_s
}

type Extensions []interface{}

func (o Extensions) GetBytes() []byte {
	byte_s := []byte{0}
	return byte_s
}

type OpData interface {
	GetBytes() []byte
	SetFee(amount uint64)
}

/*手续费*/
type Fee struct {
	FeeData Amount `json:"fee"`
}

func (o *Fee) SetFee(amount uint64) {
	o.FeeData.Amount = amount
	return
}

type UpgradeAccount struct {
	Fee
	AccountToUpgrade        ObjectId   `json:"account_to_upgrade"`
	UpgradeToLifetimeMember bool       `json:"upgrade_to_lifetime_member"`
	ExtensionsData          Extensions `json:"extensions"`
}

func CreateUpgradeAccount(name string, account_id string) *UpgradeAccount {
	//info := rpc.GetAccountInfoByName(name)
	u := &UpgradeAccount{
		ExtensionsData:          []interface{}{},
		AccountToUpgrade:        ObjectId(account_id),
		UpgradeToLifetimeMember: true,
	}
	u.FeeData = Amount{Amount: 0, AssetID: "1.3.0"}
	return u
}

func (o UpgradeAccount) GetBytes() []byte {
	fee_data := o.FeeData.GetBytes()
	atu_data := o.AccountToUpgrade.GetBytes()
	utlm_data := []byte{0}
	if o.UpgradeToLifetimeMember {
		utlm_data = []byte{1}
	}
	extensions_data := o.ExtensionsData.GetBytes()
	byte_s := append(fee_data,
		append(atu_data,
			append(utlm_data, extensions_data...)...)...)
	//fmt.Println("op byte len:::", len(byte_s))
	return byte_s
}

type KeyInfo struct {
	WeightThreshold int64         `json:"weight_threshold"`
	AccountAuths    []interface{} `json:"account_auths"`
	KeyAuths        [][]string    `json:"key_auths"`
	ExtensionsData  Extensions    `json:"extensions"`
}

func (o KeyInfo) GetBytes() []byte {
	wt_data := common.VarInt(o.WeightThreshold, 32)
	ka_data := []byte{byte(len(o.KeyAuths))}
	for i := 0; i < len(o.KeyAuths); i++ {
		tmp, _ := strconv.Atoi(o.KeyAuths[i][1])
		ka_data = append(ka_data, append(PukBytesFromBase58String(o.KeyAuths[i][0]), common.VarInt(int64(tmp), 16)...)...)
	}
	aa_data := []byte{byte(len(o.AccountAuths))}
	extensions_data := o.ExtensionsData.GetBytes()
	byte_s := append(wt_data,
		append(aa_data,
			append(ka_data, extensions_data...)...)...)
	////fmt.Println("key_info_len", len(byte_s))
	////fmt.Println("key_info:::", byte_s)
	return byte_s
}

type Options struct {
	ExtensionsData Extensions    `json:"extensions"`
	NumCommittee   int64         `json:"num_committee"`
	MemoKey        string        `json:"memo_key"`
	Votes          []interface{} `json:"votes"`
	NumWitness     int64         `json:"num_witness"`
	VotingAccount  ObjectId      `json:"voting_account"`
}

func (o Options) GetBytes() []byte {
	nc_data := common.VarInt(o.NumCommittee, 16)
	nw_data := common.VarInt(o.NumWitness, 16)
	mk_data := PukBytesFromBase58String(o.MemoKey)
	votes_data := []byte{0x00}
	va_data := o.VotingAccount.GetBytes()
	extensions_data := o.ExtensionsData.GetBytes()
	byte_s := append(mk_data,
		append(va_data,
			append(nw_data,
				append(nc_data,
					append(votes_data, extensions_data...)...)...)...)...)
	////fmt.Println("Options_len", len(byte_s))
	////fmt.Println("Options:::", byte_s)
	return byte_s
}

type String string

func (o String) GetBytes() []byte {
	byte_s := append([]byte{byte(len(o))}, []byte(o)...)
	return byte_s
}

type RegisterData struct {
	Fee
	Referrer        ObjectId   `json:"referrer"`
	ExtensionsData  Extensions `json:"extensions"`
	Active          KeyInfo    `json:"active"`
	OptionsData     Options    `json:"options"`
	Owner           KeyInfo    `json:"owner"`
	ReferrerPercent int64      `json:"referrer_percent"`
	Name            String     `json:"name"`
	Registrar       ObjectId   `json:"registrar"`
}

func (o RegisterData) GetBytes() []byte {
	fee_data := o.FeeData.GetBytes()
	referrer_data := o.Referrer.GetBytes()
	registrar_data := o.Registrar.GetBytes()
	referrer_percent_data := common.VarInt(o.ReferrerPercent, 16)
	name_data := o.Name.GetBytes()
	owner_data := o.Owner.GetBytes()
	active_data := o.Active.GetBytes()
	op_data := o.OptionsData.GetBytes()
	extensions_data := o.ExtensionsData.GetBytes()
	byte_s := append(fee_data,
		append(registrar_data,
			append(referrer_data,
				append(referrer_percent_data,
					append(name_data,
						append(owner_data,
							append(active_data,
								append(op_data, extensions_data...)...)...)...)...)...)...)...)
	return byte_s
}

func CreateRegisterData(active_PubKey, owner_PubKey, name, referrer, registrar string) *RegisterData {

	active_key := KeyInfo{
		ExtensionsData:  []interface{}{},
		AccountAuths:    []interface{}{},
		WeightThreshold: 1,
		KeyAuths:        [][]string{[]string{active_PubKey, "1"}},
	}
	owner_key := KeyInfo{
		ExtensionsData:  []interface{}{},
		AccountAuths:    []interface{}{},
		WeightThreshold: 1,
		KeyAuths:        [][]string{[]string{owner_PubKey, "1"}},
	}
	opData := Options{
		ExtensionsData: []interface{}{},
		NumWitness:     0,
		NumCommittee:   0,
		MemoKey:        active_PubKey,
		Votes:          []interface{}{},
		VotingAccount:  ObjectId(registrar),
	}
	r := &RegisterData{
		Referrer:        ObjectId(referrer),
		Registrar:       ObjectId(registrar),
		Name:            String(name),
		ExtensionsData:  []interface{}{},
		Active:          active_key,
		Owner:           owner_key,
		ReferrerPercent: 5000,
		OptionsData:     opData,
	}
	r.FeeData = Amount{Amount: 0, AssetID: "1.3.0"}
	return r
}

type CoreExchangeRate struct {
	Base  Amount `json:"base"`
	Quote Amount `json:"quote"`
}

func (o CoreExchangeRate) GetBytes() []byte {
	byte_s := append(o.Base.GetBytes(), o.Quote.GetBytes()...)
	fmt.Println("CoreExchangeRate len:::", len(byte_s))
	return byte_s
}

type CommonOptions struct {
	MaxSupply            uint64           `json:"max_supply"`
	MarketFeePercent     uint64           `json:"market_fee_percent"`
	MaxMarketFee         uint64           `json:"max_market_fee"`
	IssuerPermissions    uint64           `json:"issuer_permissions"`
	Flags                uint64           `json:"flags"`
	CoreExchangeRateData CoreExchangeRate `json:"core_exchange_rate"`
	Description          String           `json:"description"`
	Extensions           Extensions       `json:"extensions"`
}

func (o CommonOptions) GetBytes() []byte {
	MaxSupply_data := common.VarUint(o.MaxSupply, 64)
	MarketFeePercent_data := common.VarUint(o.MarketFeePercent, 16)
	MaxMarketFee_data := common.VarUint(o.MaxMarketFee, 64)
	IssuerPermissions_data := common.VarUint(o.IssuerPermissions, 16)
	Flags_data := common.VarUint(o.Flags, 16)
	CoreExchangeRate_data := o.CoreExchangeRateData.GetBytes()
	des_data := o.Description.GetBytes()
	extensions_data := o.Extensions.GetBytes()
	byte_s := append(MaxSupply_data,
		append(MarketFeePercent_data,
			append(MaxMarketFee_data,
				append(IssuerPermissions_data,
					append(Flags_data,
						append(CoreExchangeRate_data,
							append(des_data, extensions_data...)...)...)...)...)...)...)
	//fmt.Println("CommonOptions byte len:::", len(byte_s))
	return byte_s
}

/*创建代币的数据结构*/
type CreateAssetData struct {
	Fee
	Issuer            ObjectId      `json:"issuer"`
	Symbol            String        `json:"symbol"`
	Precision         uint64        `json:"precision"`
	CommonOptionsData CommonOptions `json:"common_options"`
	Extensions        Extensions    `json:"extensions"`
}

func (o CreateAssetData) GetBytes() []byte {
	fee_data := o.FeeData.GetBytes()
	issuer_data := o.Issuer.GetBytes()
	symbol_data := o.Symbol.GetBytes()
	precision_data := common.VarUint(o.Precision, 8)
	cod_data := o.CommonOptionsData.GetBytes()
	bo_data := common.VarUint(0, 8)
	extensions_data := o.Extensions.GetBytes()
	byte_s := append(fee_data,
		append(issuer_data,
			append(symbol_data,
				append(precision_data,
					append(cod_data,
						append(bo_data, extensions_data...)...)...)...)...)...)
	//fmt.Println("CreateAssetData byte len:::", len(byte_s))
	return byte_s
}

/*创建發行代币的数据结构*/
type IssueAsset struct {
	Fee
	Issuer         ObjectId   `json:"issuer"`
	AssetToIssue   Amount     `json:"asset_to_issue"`
	IssueToAccount ObjectId   `json:"issue_to_account"`
	Extensions     Extensions `json:"extensions"`
}

func (o IssueAsset) GetBytes() []byte {
	byte_s := append(o.FeeData.GetBytes(),
		append(o.Issuer.GetBytes(),
			append(o.AssetToIssue.GetBytes(),
				append(o.IssueToAccount.GetBytes(),
					append([]byte{0x0},
						o.Extensions.GetBytes()...)...)...)...)...)
	fmt.Println("IssueAsset len:::", len(byte_s))
	return byte_s
}

type Transaction struct {
	Fee
	From           ObjectId   `json:"from"`
	To             ObjectId   `json:"to"`
	AmountData     Amount     `json:"amount"`
	MemoData       Memo       `json:"memo"`
	ExtensionsData Extensions `json:"extensions"`
}

func (o Transaction) GetBytes() []byte {
	fee_data := o.FeeData.GetBytes()
	from_data := o.From.GetBytes()
	to_data := o.To.GetBytes()
	amount_data := o.AmountData.GetBytes()
	memo_data := o.MemoData.GetBytes()
	//fmt.Println("memo len:", len(memo_data))
	extensions_data := o.ExtensionsData.GetBytes()
	byte_s := append(fee_data,
		append(from_data,
			append(to_data,
				append(amount_data,
					append(memo_data, extensions_data...)...)...)...)...)
	//fmt.Println("op byte len:::", len(byte_s))
	return byte_s
}